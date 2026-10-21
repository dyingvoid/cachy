include .env
export

.PHONY: run-master run-slave run-client

export PROJECT_ROOT=$(shell pwd)

CLIENT_HOST ?= localhost
CLIENT_PORT ?= 8081

run-master:
	@go fmt ${PROJECT_ROOT}/...
	@ROLE=master SERVER_PORT=8081 go run ${PROJECT_ROOT}/cmd/server

run-slave:
	@go fmt ${PROJECT_ROOT}/...
	@ROLE=slave REPLICA_OF=localhost:8081 SERVER_PORT=8082 go run ${PROJECT_ROOT}/cmd/server

run-client:
	@go fmt ${PROJECT_ROOT}/...
	@go run ${PROJECT_ROOT}/cmd/client -h ${CLIENT_HOST} -p ${CLIENT_PORT}
