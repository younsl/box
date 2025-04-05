# Makefile for resume HTML to PDF conversion

# OS detection
UNAME_S := $(shell uname -s)

# Variables
RESUME_HTML := resume.html
OUTPUT_PDF := resume.pdf

# Default target
.PHONY: all
all: open

# OS-specific browser commands
ifeq ($(UNAME_S),Darwin)  # macOS
    BROWSER_CMD := open -a "Google Chrome"
    PRINT_CMD := /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --headless --disable-gpu --print-to-pdf-no-header --print-to-pdf=$(OUTPUT_PDF) $(RESUME_HTML)
else ifeq ($(UNAME_S),Linux)  # Linux
    BROWSER_CMD := google-chrome
    PRINT_CMD := google-chrome --headless --disable-gpu --print-to-pdf-no-header --print-to-pdf=$(OUTPUT_PDF) $(RESUME_HTML)
else  # Windows or other
    BROWSER_CMD := echo "Unsupported OS for automated commands"
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
	@echo "Generating PDF from $(RESUME_HTML)..."
	$(PRINT_CMD)
	@if [ $$? -eq 0 ]; then \
		echo "PDF generated successfully: $(OUTPUT_PDF)"; \
	else \
		echo "For manual PDF creation:"; \
		echo "1. Open $(RESUME_HTML) in Chrome"; \
		echo "2. Press Ctrl+P or Cmd+P"; \
		echo "3. Change destination to 'Save as PDF'"; \
		echo "4. Click Save"; \
		$(MAKE) open; \
	fi

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