PROJECT_NAME := goload

include api/Makefile


.PHONY: generate
generate:
	buf generate api

.PHONY: lint
lint:
	golangci-lint run ./...
