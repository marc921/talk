include .env
export

.PHONY: build
build:
	earthly +build
	earthly +run
	@docker login -u ${DOCKER_USER} --password-stdin <<< ${DOCKER_PASSWORD}
	docker push $(DOCKER_USER)/$(COMPONENT):latest