include .env
export

export PROJECT_ROOT=$(shell pwd)

run:
	@go fmt ${PROJECT_ROOT}/... && go run ${PROJECT_ROOT}/cmd