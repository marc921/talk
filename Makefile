include .env
export

REMOTE_HOST := marcbrun.eu
CLIENT_REMOTE_PATH := ~/public/talkclient
CLIENT_BINARY := talkclient

CLIENT_DATABASE_URL=sqlite3:$(HOME)/.config/talk/database.sqlite3
SERVER_DATABASE_URL=sqlite3:server_database.sqlite3
DB_CLIENT_DIR=internal/client/database
DB_SERVER_DIR=internal/server/database

.PHONY: build-server
build-server:
	earthly +build
	earthly +run
	@docker login -u ${DOCKER_USER} --password-stdin <<< ${DOCKER_PASSWORD}
	docker push ${DOCKER_USER}/talk_server:latest

.PHONY: restart-server
restart-server:
	ssh $(REMOTE_HOST) ' \
		if [ -n "$$(docker ps -a -q)" ]; then docker stop $$(docker ps -a -q) && docker rm `docker ps -a -q`; fi && \
		docker image rm ${DOCKER_USER}/talk_server || true && \
		docker run -d \
			-e AUTH_CHALLENGE_SECRET_KEY=${AUTH_CHALLENGE_SECRET_KEY} \
			-e AUTH_TOKEN_SECRET_KEY=${AUTH_TOKEN_SECRET_KEY} \
			-e DATABASE_URL=${SERVER_DATABASE_URL} \
			--network host \
			--name talk_server \
			--volume ~/public/:/bin/public/ \
			${DOCKER_USER}/talk_server \
		'

.PHONY: deploy-server
deploy-server: build-server restart-server

.PHONY: local-server
local-server:
	templ generate
	TLS=false go run ./cmd/server

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
	dbmate --url $(CLIENT_DATABASE_URL) --migrations-dir $(DB_CLIENT_DIR)/migrations --schema-file $(DB_CLIENT_DIR)/schema.sql drop
	dbmate --url $(CLIENT_DATABASE_URL) --migrations-dir $(DB_CLIENT_DIR)/migrations --schema-file $(DB_CLIENT_DIR)/schema.sql up

.PHONY: recreate-server-db
recreate-server-db:
	dbmate --url $(SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations --schema-file $(DB_SERVER_DIR)/schema.sql drop
	dbmate --url $(SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations --schema-file $(DB_SERVER_DIR)/schema.sql up

.PHONY: server-db-add-migration
server-db-add-migration:
	@if [ -z "$$name" ]; then \
		echo "Error: Migration name is required. Usage: make server-db-add-migration name=<migration_name>"; \
		exit 1; \
	fi
	dbmate --url $(SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations new $$name

.PHONY: generate
generate:	# (Re)Generate automatically generated code, including sqlc queries and openapi client
	go generate ./...

.PHONY: prune
prune:	# Prune docker and earthly cache
	docker system prune -a
	earthly prune