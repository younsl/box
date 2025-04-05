# OS detection
UNAME_S := $(shell uname -s)

# Default target
.PHONY: all
all: open

# Open resume.html
.PHONY: open
open:
ifeq ($(UNAME_S),Darwin)  # macOS
	open -a "Google Chrome" resume.html
else ifeq ($(UNAME_S),Linux)  # Linux
	if command -v google-chrome > /dev/null; then \
		google-chrome resume.html; \
	elif command -v chromium-browser > /dev/null; then \
		chromium-browser resume.html; \
	else \
		echo "Chrome or Chromium browser not found."; \
	fi
else
	echo "Unsupported operating system."
endif

# Export to PDF (additional feature)
.PHONY: pdf
pdf:
	@echo "To export as PDF, open resume.html in Chrome and press Ctrl+P or Cmd+P to print, then select 'Save as PDF'."
	$(MAKE) open

# Help
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make - Open resume.html in Chrome"
	@echo "  make open - Open resume.html in Chrome"
	@echo "  make pdf - Show PDF export instructions and open in Chrome"
	@echo "  make help - Display help"

# Failure message
.PHONY: failed
failed:
	@echo "Command execution failed. Please make sure Chrome is installed."