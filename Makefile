-include .env
export

COMPOSE ?= docker compose

.PHONY: run-master run-slave run-client-master run-client-slave \
	docker-build docker-up docker-down docker-ps docker-logs \
	docker-logs-master docker-logs-slave \
	docker-client-master docker-client-slave

run-master:
	@go fmt ./...
	@ROLE=master SERVER_PORT=8082 go run ./cmd/server

run-slave:
	@go fmt ./...
	@REPLICATION_ROLE=slave SERVER_PORT=8083 go run ./cmd/server

run-client-master:
	@go fmt ./...
	@CLIENT_PORT=8083 go run ./cmd/client

run-client-slave:
	@go fmt ./...
	@CLIENT_PORT=8083 go run ./cmd/client

docker-build:
	@$(COMPOSE) --profile clients build

docker-up: docker-build
	@$(COMPOSE) up -d --wait

docker-down:
	@$(COMPOSE) down

docker-client-master:
	@$(COMPOSE) --profile clients build client-master
	@$(COMPOSE) --profile clients run --rm client-master

docker-client-slave:
	@$(COMPOSE) --profile clients build client-slave
	@$(COMPOSE) --profile clients run --rm client-slave
