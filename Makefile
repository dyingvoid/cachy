-include .env
export

.PHONY: run-master run-slave run-client

run-master:
	@go fmt ./...
	@ROLE=master SERVER_PORT=8081 go run ./cmd/server

run-slave:
	@go fmt ./...
	@ROLE=slave REPLICA_OF=${CLIENT_HOST}:${CLIENT_PORT} SERVER_PORT=8082 go run ./cmd/server

run-client:
	@go fmt ./...
	@go run ./cmd/client
