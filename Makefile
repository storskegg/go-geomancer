.PHONY: all build test lint fmt vet tidy

all: lint test build

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
