# hago Makefile

BINARY := hago
MODULE := github.com/rmrfslashbin/hago

# Build information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build build-all test test-cover lint lint-fix vet fmt clean help

## build: Build the CLI binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

## build-all: Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 ./cmd/$(BINARY)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/$(BINARY)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

## test: Run tests
test:
	go test -v -race ./...

## test-cover: Run tests with coverage
test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint
lint:
	golangci-lint run

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	golangci-lint run --fix

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## tidy: Tidy go modules
tidy:
	go mod tidy

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## clean: Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## install: Install the binary locally
install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

.DEFAULT_GOAL := help
