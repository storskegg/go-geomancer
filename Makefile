.PHONY: all build geomancer geomancer-export test test-full cover golden lint fmt vet tidy

all: lint test build

build:
	go build ./...

geomancer:
	go build -o bin/geomancer ./cmd/geomancer

geomancer-export:
	go build -o bin/geomancer-export ./cmd/geomancer-export

# -short skips the full-size export tests, which are slow under -race.
test:
	go test -race -short ./...

# Every test, including full exports in each format (without -race).
test-full:
	go test ./...

# Per-function coverage, least-covered last.
cover:
	go test -short -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -k3 -rn

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
	go generate ./internal/licenses
