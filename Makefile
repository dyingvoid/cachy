include .env
export

export PROJECT_ROOT=$(shell pwd)

run:
	@go run ${PROJECT_ROOT}/cmd