# Resume HTML Management Makefile
# ================================

# Configuration
RESUME_FILE := resume.html
PDF_PATTERN := *.pdf

# OS-specific browser detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    BROWSER := open -a "Google Chrome"
else ifeq ($(UNAME_S),Linux)
    BROWSER := google-chrome
else
    BROWSER := echo "Unsupported OS for automated browser opening"
endif

# Targets
.PHONY: all open clean help

# Default target
all: open

# Open resume in browser
open:
	@echo "Opening $(RESUME_FILE) in Chrome..."
	@$(BROWSER) $(RESUME_FILE)

# Clean generated files
clean:
	@echo "Cleaning generated PDF files..."
	@rm -f $(PDF_PATTERN)
	@echo "Clean completed."

# Show available commands
help:
	@echo ""
	@echo "Resume HTML Makefile"
	@echo "===================="
	@echo ""
	@echo "Available commands:"
	@echo "  make         - Open resume in Chrome"
	@echo "  make open    - Open resume in Chrome"
	@echo "  make clean   - Remove generated PDF files"
	@echo "  make help    - Show this help message"
	@echo ""
