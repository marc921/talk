include .env
export

REMOTE_HOST := marcbrun.eu
CLIENT_REMOTE_PATH := ~/public/talkclient
CLIENT_BINARY := talkclient

.PHONY: build-server
build-server:
	earthly +build
	earthly +run
	@docker login -u ${DOCKER_USER} --password-stdin <<< ${DOCKER_PASSWORD}
	docker push ${DOCKER_USER}/talk_server:latest

.PHONY: restart-server
restart-server:
	ssh $(REMOTE_HOST) ' \
		docker stop `docker ps -a -q` && \
		docker rm `docker ps -a -q` && \
		docker image rm ${DOCKER_USER}/talk_server && \
		docker run -d \
			-e AUTH_CHALLENGE_SECRET_KEY=${AUTH_CHALLENGE_SECRET_KEY} \
			-e AUTH_TOKEN_SECRET_KEY=${AUTH_TOKEN_SECRET_KEY} \
			--network host \
			--name talk_server \
			--volume ~/public/:/bin/public/ \
			${DOCKER_USER}/talk_server \
		'

.PHONY: deploy-server
deploy-server: build-server restart-server

.PHONY: build-client
build-client:
	CGO_ENABLED=1 go build -o $(CLIENT_BINARY) ./cmd/client

.PHONY: push-client
push-client: build-client
	@echo "Deploying to $(REMOTE_HOST)..."
	rsync -avz --progress $(CLIENT_BINARY) $(REMOTE_HOST):$(CLIENT_REMOTE_PATH)
	@if [ $$? -eq 0 ]; then \
		echo "Deployment successful"; \
		ssh $(REMOTE_HOST) 'chmod +x $(CLIENT_REMOTE_PATH)'; \
		rm $(CLIENT_BINARY); \
	else \
		echo "Deployment failed"; \
		rm $(CLIENT_BINARY); \
		exit 1; \
	fi

.PHONY: deploy-all
deploy-all: deploy-server push-client

.PHONY: recreate-client-db
recreate-client-db:
	dbmate drop
	dbmate up

.PHONY: generate
generate:	# (Re)Generate automatically generated code, including sqlc queries and openapi client
	go generate ./...
