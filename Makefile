include .env
export

.PHONY: build
build:
	earthly +build
	earthly +run
	@docker login -u ${DOCKER_USER} --password-stdin <<< ${DOCKER_PASSWORD}
	docker push $(DOCKER_USER)/$(COMPONENT):latest

.PHONY: recreate-client-db
recreate-client-db:
	dbmate drop
	dbmate up

.PHONY: generate
generate:	# (Re)Generate automatically generated code, including sqlc queries and openapi client
	go generate ./...