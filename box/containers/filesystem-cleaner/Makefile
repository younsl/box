BINARY_NAME := filesystem-cleaner
# Use environment variables if set, otherwise fallback to git/system values
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Docker/ECR settings
DOCKER_IMAGE := filesystem-cleaner
DOCKER_TAG := $(VERSION)
ECR_REGISTRY := your-account.dkr.ecr.us-east-1.amazonaws.com
ECR_IMAGE := $(ECR_REGISTRY)/$(DOCKER_IMAGE)

.PHONY: all
all: deps fmt vet test build

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)

.PHONY: run
run: build
	./bin/$(BINARY_NAME) --dry-run --threshold 70 --paths /tmp

.PHONY: dev
dev: build
	./bin/$(BINARY_NAME) --log-level debug --dry-run --threshold 70 --paths /tmp

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: docker-build
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(ECR_REGISTRY)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(ECR_IMAGE):$(DOCKER_TAG)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(ECR_IMAGE):latest
	docker push $(ECR_IMAGE):$(DOCKER_TAG)
	docker push $(ECR_IMAGE):latest

.PHONY: install
install: build
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/

.PHONY: uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Run deps, fmt, vet, test, and build"
	@echo "  build       - Build the binary"
	@echo "  build-all   - Build for all platforms"
	@echo "  run         - Build and run with default settings"
	@echo "  dev         - Run with debug logging"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run golangci-lint"
	@echo "  deps        - Update dependencies"
	@echo "  clean       - Remove build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push - Push to ECR"
	@echo "  install     - Install to /usr/local/bin"
	@echo "  uninstall   - Remove from /usr/local/bin"