VERSION 0.8

# Target to build the Go project
build:
	FROM golang:latest
	WORKDIR /app

	COPY go.mod go.sum ./
	RUN go mod tidy
	RUN --mount=type=cache,target=/go/pkg/mod go mod download

	COPY . .
	RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o talkserver ./cmd/server
	SAVE ARTIFACT talkserver

# Target to run the built binary
run:
	FROM alpine:latest
	WORKDIR /bin
	# poppler-utils is required for PDF text extraction
	RUN apk add --no-cache poppler-utils
	COPY +build/talkserver /bin/talkserver
	COPY ./server_database.sqlite3 /bin/server_database.sqlite3
	EXPOSE 443
	ENTRYPOINT ["./talkserver"]
	SAVE IMAGE marcbrunlearning/talk_server:latest