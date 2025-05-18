include .env
export

REMOTE_HOST := marcbrun.eu
CLIENT_REMOTE_PATH := ~/public/talkclient
CLIENT_BINARY := talkclient

CLIENT_DATABASE_URL=sqlite3:$(HOME)/.config/talk/database.sqlite3
SERVER_DATABASE_URL=postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable
TUNNEL_SERVER_DATABASE_URL=postgres://myuser:mypassword@localhost:15432/mydb?sslmode=disable
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
		if docker ps -a --format "{{.Names}}" | grep -q "^talk_server$$"; then \
			docker stop talk_server && docker rm talk_server; \
		fi && \
		docker image rm ${DOCKER_USER}/talk_server || true && \
		docker run -d \
			-e AUTH_CHALLENGE_SECRET_KEY=${AUTH_CHALLENGE_SECRET_KEY} \
			-e AUTH_TOKEN_SECRET_KEY=${AUTH_TOKEN_SECRET_KEY} \
			-e DATABASE_URL=${SERVER_DATABASE_URL} \
			--network host \
			--name talk_server \
			--volume ~/public/:/bin/public/ \
			--volume ~/talk_tls_cache:/var/www/.cache \
			${DOCKER_USER}/talk_server \
		'
	@printf "${GREEN}✅ Server restarted successfully${NC}\n"

.PHONY: deploy-server
deploy-server: build-server restart-server

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
	ssh -L 15432:localhost:5432 $(REMOTE_HOST) -N & \
	TUNNEL_PID=$$!; \
	echo "SSH tunnel started with PID $$TUNNEL_PID"; \
	sleep 1; \
	dbmate --url $(TUNNEL_SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations --schema-file $(DB_SERVER_DIR)/schema.sql drop; \
	dbmate --url $(TUNNEL_SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations --schema-file $(DB_SERVER_DIR)/schema.sql up; \
	kill $$TUNNEL_PID

.PHONY: server-db-add-migration
server-db-add-migration:
	@if [ -z "$$name" ]; then \
		echo "Error: Migration name is required. Usage: make server-db-add-migration name=<migration_name>"; \
		exit 1; \
	fi; \
	ssh -L 15432:localhost:5432 $(REMOTE_HOST) -N & \
	TUNNEL_PID=$$!; \
	echo "SSH tunnel started with PID $$TUNNEL_PID"; \
	sleep 1; \
	dbmate --url $(TUNNEL_SERVER_DATABASE_URL) --migrations-dir $(DB_SERVER_DIR)/migrations new $$name; \
	kill $$TUNNEL_PID

server-db-connect:
	ssh -L 15432:localhost:5432 $(REMOTE_HOST) -N & \
	TUNNEL_PID=$$!; \
	echo "SSH tunnel started with PID $$TUNNEL_PID"; \
	sleep 1; \
	PGPASSWORD=mypassword psql -h localhost -p 15432 -U myuser -d mydb; \
	kill $$TUNNEL_PID

server-db-create:
	ssh $(REMOTE_HOST) ' \
		docker run \
			--name postgres \
			-e POSTGRES_PASSWORD=mypassword \
			-e POSTGRES_USER=myuser \
			-e POSTGRES_DB=mydb \
			-p 5432:5432 \
			-v postgres_data:/var/lib/postgresql/data \
			-d postgres; \
	'

server-container-shell:
	ssh -t $(REMOTE_HOST) ' \
		docker exec -it talk_server -- sh; \
	'

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
	cd cmd/server/frontend && REACT_APP_API_URL=https://marcbrun.eu/api/v1 npm run build


.PHONY: local-server
local-server:
	ssh -L 15432:localhost:5432 $(REMOTE_HOST) -N & \
	TUNNEL_PID=$$!; \
	echo "SSH tunnel started with PID $$TUNNEL_PID"; \
	sleep 1; \
	DATABASE_URL=${TUNNEL_SERVER_DATABASE_URL} TLS=false go run ./cmd/server; \
	kill $$TUNNEL_PID

local-frontend:
	cd cmd/server/frontend && REACT_APP_API_URL=http://localhost:8080/api/v1 npm start

local-client:
	go run ./cmd/client

reset-db: recreate-client-db recreate-server-db generate