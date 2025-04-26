# Output binary name
BINARY_NAME=ol
# Output directory for the binary
OUTPUT_DIR=./bin
# Build path for the main package
CMD_PATH=./cmd/ol
# Go build flags
BUILD_FLAGS=-tags timetzdata

# Default target: build the application
all: build

# Build the Go application
build:
	@echo "Building ${BINARY_NAME} to ${OUTPUT_DIR}..."
	@mkdir -p ${OUTPUT_DIR}
	@go build ${BUILD_FLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME} ${CMD_PATH}
	@echo "${BINARY_NAME} built successfully in ${OUTPUT_DIR}/."

# Remove the built binary and output directory
clean:
	@echo "Cleaning up..."
	@rm -rf ${OUTPUT_DIR}
	@echo "Cleanup complete."

# Install the application using go install
install:
	@echo "Installing ${BINARY_NAME}..."
	@go install ${BUILD_FLAGS} ${CMD_PATH}
	@echo "${BINARY_NAME} installed successfully."

# Run tests (if any) - Placeholder
test:
	@echo "Running tests..."
	@go test ./...

.PHONY: all build clean install test 