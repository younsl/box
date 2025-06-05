.PHONY: build clean test run help

BINARY_NAME=jdk-scanner
BUILD_DIR=bin
CMD_DIR=cmd/scanner

help: ## 도움말 출력
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## 바이너리 빌드
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## 빌드 아티팩트 정리
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean

test: ## 테스트 실행
	@echo "Running tests..."
	@go test -v ./...

run: build ## 빌드 후 실행
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

install: build ## 시스템에 설치
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

fmt: ## 코드 포맷팅
	@echo "Formatting code..."
	@go fmt ./...

lint: ## 코드 린팅 (golangci-lint 필요)
	@echo "Running linter..."
	@golangci-lint run

mod: ## 모듈 정리
	@echo "Cleaning up modules..."
	@go mod tidy
	@go mod verify 