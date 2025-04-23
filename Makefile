include .env
export

REMOTE_HOST := marcbrun.eu
CLIENT_REMOTE_PATH := ~/public/talkclient
CLIENT_BINARY := talkclient

CLIENT_DATABASE_URL=sqlite3:$(HOME)/.config/talk/database.sqlite3
SERVER_DATABASE_URL=sqlite3:server_database.sqlite3
DB_CLIENT_DIR=internal/client/database
DB_SERVER_DIR=internal/server/database

# Define some color variables
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: build-server
build-server: build-react
	@printf "${BLUE}🏗️ Building server...${NC}\n"
	earthly +build
	earthly +run
	@docker login -u ${DOCKER_USER} --password-stdin <<< ${DOCKER_PASSWORD}
	docker push ${DOCKER_USER}/talk_server:latest
	@printf "${GREEN}✅ Server built successfully${NC}\n"

.PHONY: restart-server
restart-server:
	@printf "${BLUE}🔄 Restarting server...${NC}\n"
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
	@printf "${GREEN}✅ Server restarted successfully${NC}\n"

.PHONY: deploy-server
deploy-server: build-server restart-server

.PHONY: local-server
local-server:
	DATABASE_URL=server_database.sqlite3 TLS=false go run ./cmd/server

.PHONY: build-client
build-client:
	@printf "${BLUE}🏗️ Building client...${NC}\n"
	CGO_ENABLED=1 go build -o $(CLIENT_BINARY) ./cmd/client
	@printf "${GREEN}✅ Client built successfully${NC}\n"

.PHONY: push-client
push-client: build-client
	@printf "${BLUE}🚀 Pushing client to remote host $(REMOTE_HOST)...${NC}\n"
	rsync -avz --progress $(CLIENT_BINARY) $(REMOTE_HOST):$(CLIENT_REMOTE_PATH)
	@if [ $$? -eq 0 ]; then \
		printf "${GREEN}✅ Client pushed successfully${NC}\n"; \
		ssh $(REMOTE_HOST) 'chmod +x $(CLIENT_REMOTE_PATH)'; \
		rm $(CLIENT_BINARY); \
	else \
		printf "${RED}❌ Error pushing client${NC}\n"; \
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

.PHONY: build-react
build-react:
	@printf "${BLUE}📦 Installing npm dependencies...${NC}\n"
	cd cmd/server/frontend && npm install
	@printf "${BLUE}🔍 Type checking TypeScript...${NC}\n"
	cd cmd/server/frontend && npx tsc --noEmit
	@printf "${BLUE}💅 Building Tailwind CSS...${NC}\n"
	cd cmd/server/frontend && npx tailwindcss -i ./src/App.css -o ./src/tailwind.css
	@printf "${BLUE}⚛️ Building React app...${NC}\n"
	cd cmd/server/frontend && npm run build

local-frontend:
	cd cmd/server/frontend && npm start