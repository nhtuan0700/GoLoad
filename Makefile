RUN_GO=docker compose exec go
RUN_SWAG=docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli:v7.6.0
RUN_BUF=docker run --rm -v "${PWD}:/workspace" --workdir /workspace bufbuild/buf:1.34.0

.PHONY: tidy
tidy:
	$(RUN_GO) go mod tidy

.PHONY: generate
generate:
	$(RUN_GO) protoc -I=. \
		--go_out=internal/generated \
		--go-grpc_out=internal/generated \
		--grpc-gateway_out=internal/generated \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out .\
		--openapiv2_opt generate_unbound_methods=true \
		--validate_out="lang=go:internal/generated" \
		api/go_load.proto

	$(RUN_GO) wire internal/wiring/wire.go
	make tidy

.PHONY: standalone-server
standalone-server:
	$(RUN_GO) go run cmd/*.go standalone-server

.PHONY: lint
lint:
	$(RUN_GO) golangci-lint run ./...

.PHONY: generate-swagger
generate-swagger:
	$(RUN_SWAG) generate -i /local/api/go_load.swagger.json -g typescript-fetch -o /local/output/client/$(APP)
