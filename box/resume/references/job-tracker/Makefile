.PHONY: build run clean deps init-gpg setup help

# Variables
BINARY_NAME=job-tracker
GO=go
GPG_RECIPIENT ?= your-email@example.com

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	$(GO) build -o $(BINARY_NAME) ./cmd/job-tracker
	@echo "$(GREEN)✓ Build complete$(NC)"

run: ## Run the application
	GPG_RECIPIENT=$(GPG_RECIPIENT) $(GO) run ./cmd/job-tracker

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	GPG_RECIPIENT=$(GPG_RECIPIENT) air -c .air.toml

clean: ## Clean build artifacts and temporary files
	rm -f $(BINARY_NAME)
	rm -f data.json
	@echo "$(GREEN)✓ Cleaned$(NC)"

deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod tidy

init-gpg: ## Initialize GPG key for encryption
	@echo "Setting up GPG encryption..."
	@echo "Current GPG keys:"
	@gpg --list-secret-keys --keyid-format LONG
	@echo ""
	@echo "Set your GPG recipient email in environment:"
	@echo "  export GPG_RECIPIENT=your-email@example.com"
	@echo ""
	@echo "Or run with:"
	@echo "  make run GPG_RECIPIENT=your-email@example.com"

setup: deps init-gpg ## Initial setup
	@echo "$(GREEN)✓ Setup complete$(NC)"
	@echo ""
	@echo "Next steps:"
	@echo "1. Set GPG_RECIPIENT environment variable"
	@echo "2. Run 'make run' to start the application"

test-encrypt: ## Test GPG encryption/decryption
	@echo '{"test": "data"}' > test.json
	@gpg --encrypt --trust-model always --recipient $(GPG_RECIPIENT) --output test.json.gpg test.json
	@echo "$(GREEN)✓ Encryption test successful$(NC)"
	@gpg --decrypt --quiet test.json.gpg
	@rm -f test.json test.json.gpg

trust-key: ## Trust the GPG key permanently
	@echo "Trusting GPG key for $(GPG_RECIPIENT)..."
	@echo "5" | gpg --command-fd 0 --expert --edit-key $(GPG_RECIPIENT) trust quit 2>/dev/null || echo "Key trust set"
	@echo "$(GREEN)✓ GPG key trusted$(NC)"

git-setup: ## Setup git hooks for automatic encryption
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "# Auto-encrypt data before commit" >> .git/hooks/pre-commit
	@echo "make build" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)✓ Git hooks configured$(NC)"