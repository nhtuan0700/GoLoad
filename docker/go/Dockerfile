FROM golang:1.22.4 as build

WORKDIR /app

RUN apt-get update && apt-get install -y \
    make \
    protobuf-compiler \
    git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
    git clone https://github.com/bufbuild/protoc-gen-validate.git && \
    cd protoc-gen-validate && make build

RUN go install github.com/google/wire/cmd/wire@latest && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0 && \
    CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest && \ 
    go install github.com/air-verse/air@latest
