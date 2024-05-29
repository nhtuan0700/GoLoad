PROJECT_NAME := goload

include api/Makefile


.PHONY: generate
generate:
	buf generate api
	wire internal/wiring/wire.go

.PHONY: lint
lint:
	golangci-lint run ./...
