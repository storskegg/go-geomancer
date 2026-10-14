.PHONY: all build geomancer test golden lint fmt vet tidy

all: lint test build

build:
	go build ./...

geomancer:
	go build -o bin/geomancer ./cmd/geomancer

test:
	go test -race ./...

# Rewrite the CLI golden files after an intended output change.
golden:
	go test ./internal/cli/ -update

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
	go mod vendor
