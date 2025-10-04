BINARY_NAME=cocd
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/cocd/main.go

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 cmd/cocd/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 cmd/cocd/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 cmd/cocd/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/cocd/main.go

.PHONY: run
run:
	go run cmd/cocd/main.go

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: mod
mod:
	go mod tidy
	go mod vendor

.PHONY: install
install:
	go install $(LDFLAGS) cmd/cocd/main.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build binaries for all platforms"
	@echo "  run        - Run the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  fmt        - Format code"
	@echo "  mod        - Tidy and vendor modules"
	@echo "  install    - Install the binary"
	@echo "  help       - Show this help"