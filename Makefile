# hago Makefile

BINARY := hago
MODULE := github.com/rmrfslashbin/hago

# Build information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build build-all test test-cover cover-report lint lint-fix vet staticcheck deadcode vulncheck fmt tidy check check-all clean install help

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
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## cover-report: Generate HTML coverage report
cover-report: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## vet: Run go vet
vet:
	go vet ./...

## staticcheck: Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck:
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

## deadcode: Run deadcode analysis (install: go install golang.org/x/tools/cmd/deadcode@latest)
deadcode:
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	deadcode -test ./...

## vulncheck: Check for known vulnerabilities (install: go install golang.org/x/vuln/cmd/govulncheck@latest)
vulncheck:
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	golangci-lint run --fix

## fmt: Format code
fmt:
	go fmt ./...
	@which goimports > /dev/null && goimports -w . || true

## tidy: Tidy go modules
tidy:
	go mod tidy

## check: Run quick checks (fmt, vet, test)
check: fmt vet test

## check-all: Run all checks including static analysis (fmt, vet, staticcheck, deadcode, vulncheck, test-cover)
check-all: fmt vet staticcheck deadcode vulncheck test-cover
	@echo ""
	@echo "All checks passed!"

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
