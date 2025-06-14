# Makefile for resume HTML to PDF conversion

# OS detection
UNAME_S := $(shell uname -s)

# Variables
RESUME := resume.html

# Default target
.PHONY: all
all: open

# OS-specific browser commands and Chrome paths
ifeq ($(UNAME_S),Darwin)  # macOS
    BROWSER_CMD := open -a "Google Chrome"
    CHROME_PATH := /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome
    PRINT_CMD := $(CHROME_PATH) --headless --disable-gpu --print-to-pdf-no-header
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
	@echo "Opening $(RESUME) in Chrome ..."
	$(BROWSER_CMD) $(RESUME)

# Generate PDF in English
.PHONY: pdf-en
pdf-en:
	@echo "Generating English PDF ..."
	@temp_file=$$(mktemp).html; \
	sed -e 's/switchLanguage('\''en'\'')/switchLanguage('\''en'\''); document.querySelectorAll('\''.lang-btn'\'').forEach(btn => btn.style.display = '\''none'\'');/' \
	    -e 's|href="style.css"|href="file://$(PWD)/style.css"|' \
	    -e 's|</head>|<style>.lang-en { display: inline !important; } .lang-ko { display: none !important; }</style></head>|' $(RESUME) > $$temp_file; \
	$(CHROME_PATH) --headless --disable-gpu --print-to-pdf="resume-en.pdf" "file://$$temp_file" --no-pdf-header-footer && \
	rm $$temp_file && \
	echo "English PDF generated: resume-en.pdf" || \
	echo "Failed to generate English PDF"

# Generate PDF in Korean
.PHONY: pdf-ko
pdf-ko:
	@echo "Generating Korean PDF ..."
	@temp_file=$$(mktemp).html; \
	sed -e 's/switchLanguage('\''en'\'')/switchLanguage('\''ko'\''); document.querySelectorAll('\''.lang-btn'\'').forEach(btn => btn.style.display = '\''none'\'');/' \
	    -e 's|href="style.css"|href="file://$(PWD)/style.css"|' \
	    -e 's|</head>|<style>.lang-en { display: none !important; } .lang-ko { display: inline !important; }</style></head>|' $(RESUME) > $$temp_file; \
	$(CHROME_PATH) --headless --disable-gpu --print-to-pdf="resume-ko.pdf" "file://$$temp_file" --no-pdf-header-footer && \
	rm $$temp_file && \
	echo "Korean PDF generated: resume-ko.pdf" || \
	echo "Failed to generate Korean PDF"

# Generate both language PDFs
.PHONY: pdf
pdf: pdf-en pdf-ko

# Clean generated files
.PHONY: clean
clean:
	@echo "Removing generated PDF files ..."
	rm -f *.pdf

# Show help
.PHONY: help
help:
	@echo "Resume HTML Makefile"
	@echo "--------------------"
	@echo "Available targets:"
	@echo "  make         - Open resume in Chrome"
	@echo "  make open    - Open resume in Chrome"
	@echo "  make pdf     - Generate PDF from HTML file (requires Chrome)"
	@echo "  make pdf-en  - Generate English PDF (resume-en.pdf)"
	@echo "  make pdf-ko  - Generate Korean PDF (resume-ko.pdf)"
	@echo "  make clean   - Remove generated PDF files"
	@echo "  make help    - Show this help message"
