VERSION 0.8

# Target to build the Go project
build:
	FROM golang:latest
	WORKDIR /app
	COPY . .
	RUN go mod tidy
	RUN go mod download
	RUN CGO_ENABLED=0 GOOS=linux go build -o talkserver ./cmd/server
	SAVE ARTIFACT talkserver

# Target to run the built binary
run:
	FROM alpine:latest
	WORKDIR /bin
	COPY +build/talkserver /bin/talkserver
	EXPOSE 443
	ENTRYPOINT ["./talkserver"]
	SAVE IMAGE marcbrunlearning/talk_server:latest