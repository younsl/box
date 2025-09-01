# Binary name
BINARY_NAME=kk

# Main package path
CMD_PATH=./cmd/kk

# Default target: build
all: build

# Build the application
build:
	@echo "Tidying $(BINARY_NAME) ..."
	@go mod tidy
	@echo "$(BINARY_NAME) tidied."
	
	@echo "Building $(BINARY_NAME) ..."
	@go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully."

# Optimized release build (strip debug info, reduce size)
release:
	@echo "Building $(BINARY_NAME) for release ..."
	@go build -ldflags="-s -w" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully (release)."

# Remove built binary
clean:
	@echo "Cleaning $(BINARY_NAME) ..."
	@rm -f $(BINARY_NAME)
	@echo "$(BINARY_NAME) cleaned."

# Run tests (for future use when tests are added)
test:
	@echo "Running tests ..."
	@go test ./... -v

# Define phony targets (.PHONY)
.PHONY: all build release clean test
