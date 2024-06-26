RUN_GO=docker compose exec go

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
