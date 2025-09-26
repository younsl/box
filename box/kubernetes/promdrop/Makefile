# Go parameters
GO := go
GOFLAGS := $(GOFLAGS)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# Build parameters
CMD_DIR := ./cmd/promdrop
BINARY_NAME := promdrop
OUTPUT_DIR := dist

# Platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Default target
all: build

# Build the application for current platform
build:
	@echo "Tidying up ..."
	go mod tidy
	@echo "Building $(BINARY_NAME) ..."
	$(GO) build $(GOFLAGS) -ldflags=$(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Cross-compile for all platforms
build-all: clean-dist
	@echo "Building for all platforms ..."
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH \
		$(GO) build $(GOFLAGS) -ldflags=$(LDFLAGS) \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH $(CMD_DIR); \
	done
	@echo "Cross-compilation complete"

# Build for specific platform (usage: make build-platform GOOS=linux GOARCH=amd64)
build-platform:
	@echo "Building for $(GOOS)/$(GOARCH) ..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build $(GOFLAGS) -ldflags=$(LDFLAGS) \
	-o $(OUTPUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) $(CMD_DIR)

# Create release archives
release-archives: build-all
	@echo "Creating release archives ..."
	@cd $(OUTPUT_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			tar -czf "$$binary.tar.gz" "$$binary"; \
			echo "Created $$binary.tar.gz"; \
		fi; \
	done

# Format the code
fmt:
	@echo "Formatting code ..."
	gofmt -w .

# Run tests
test:
	@echo "Running tests ..."
	$(GO) test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning up ..."
	rm -f $(BINARY_NAME)

# Clean distribution directory
clean-dist:
	@echo "Cleaning distribution directory ..."
	rm -rf $(OUTPUT_DIR)

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build for current platform"
	@echo "  build-all        - Cross-compile for all platforms"
	@echo "  build-platform   - Build for specific platform (set GOOS/GOARCH)"
	@echo "  release-archives - Create tar.gz archives"
	@echo "  fmt              - Format code"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean binary"
	@echo "  clean-dist       - Clean distribution directory"
	@echo "  help             - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION          - Version to embed (default: dev)"
	@echo "  COMMIT           - Git commit to embed passed as ldflags (default: unknown)"
	@echo "  GOOS             - Target OS for build-platform"
	@echo "  GOARCH           - Target architecture for build-platform"

# Declare phony targets
.PHONY: all build build-all build-platform release-archives fmt test clean clean-dist help