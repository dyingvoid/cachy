include .env
export

.PHONY: run-master run-slave

export PROJECT_ROOT=$(shell pwd)

run-master:
	@go fmt ${PROJECT_ROOT}/...
	@ROLE=master SERVER_PORT=8081 go run ${PROJECT_ROOT}/cmd

run-slave:
	@go fmt ${PROJECT_ROOT}/...
	@ROLE=slave REPLICA_OF=localhost:8081 SERVER_PORT=8082 go run ${PROJECT_ROOT}/cmd