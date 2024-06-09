FROM golang:1.22.4

WORKDIR /var/www

RUN apt-get update && apt-get install -y \
    make \
    protobuf-compiler \
    git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

RUN git clone https://github.com/bufbuild/protoc-gen-validate.git
RUN cd protoc-gen-validate && make build

RUN go install github.com/google/wire/cmd/wire@latest
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0
