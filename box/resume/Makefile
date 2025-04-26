# Makefile for resume HTML to PDF conversion

# OS detection
UNAME_S := $(shell uname -s)

# Variables
RESUME_HTML := resume-short.html
OUTPUT_PDF := resume-short.pdf

# Default target
.PHONY: all
all: open

# OS-specific browser commands and Chrome paths
ifeq ($(UNAME_S),Darwin)  # macOS
    BROWSER_CMD := open -a "Google Chrome"
    CHROME_PATH := /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome
    PRINT_CMD := $(CHROME_PATH) --headless --disable-gpu --print-to-pdf-no-header --print-to-pdf=$(OUTPUT_PDF) $(RESUME_HTML)
else ifeq ($(UNAME_S),Linux)  # Linux
    BROWSER_CMD := google-chrome
    CHROME_PATH := google-chrome
    PRINT_CMD := $(CHROME_PATH) --headless --disable-gpu --print-to-pdf-no-header --print-to-pdf=$(OUTPUT_PDF) $(RESUME_HTML)
else  # Windows or other
    BROWSER_CMD := echo "Unsupported OS for automated commands"
    CHROME_PATH := echo "Unsupported OS for automated commands"
    PRINT_CMD := echo "Unsupported OS for automated commands"
endif

# Open resume in Chrome
.PHONY: open
open:
	@echo "Opening $(RESUME_HTML) in Chrome..."
	$(BROWSER_CMD) $(RESUME_HTML)

# Generate PDF (headless Chrome)
.PHONY: pdf
pdf:
	@echo "Generating PDFs from all HTML files..."
	@for html_file in $$(find . -name "*.html"); do \
		pdf_file=$${html_file%.html}.pdf; \
		echo "Processing $$html_file -> $$pdf_file"; \
		$(CHROME_PATH) --headless --disable-gpu --print-to-pdf="$$pdf_file" "file://$(PWD)/$$html_file" --no-pdf-header-footer || \
		( \
			echo "Failed to generate PDF for $$html_file"; \
			echo "For manual PDF creation:"; \
			echo "1. Open $$html_file in Chrome"; \
			echo "2. Press Ctrl+P or Cmd+P"; \
			echo "3. Change destination to 'Save as PDF'"; \
			echo "4. Click Save"; \
		); \
	done
	@echo "PDF generation completed"

# Clean generated files
.PHONY: clean
clean:
	@echo "Removing generated PDF files..."
	rm -f *.pdf

# Show help
.PHONY: help
help:
	@echo "Resume HTML Makefile"
	@echo "--------------------"
	@echo "Available targets:"
	@echo "  make         - Open resume.html in Chrome"
	@echo "  make open    - Open resume.html in Chrome"
	@echo "  make pdf     - Generate PDF from resume.html (requires Chrome)"
	@echo "  make clean   - Remove generated PDF files"
	@echo "  make help    - Show this help message"