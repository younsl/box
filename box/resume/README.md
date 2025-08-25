# Resume

A bilingual HTML resume that supports both English and Korean languages with seamless browser-based PDF generation.

## Usage

Open the resume in your browser and generate PDFs directly:

```bash
make        # Open resume in Chrome
make clean  # Remove generated PDF files
make help   # Show available commands
```

Once opened, use the language toggle button to switch between English and Korean, then click the PDF button to generate a properly formatted PDF with automatic filename (`younsung-lee-sre-{lang}-{YYYYMMDD}.pdf`).

## Features

- **Language Toggle**: Switch between English and Korean content instantly
- **Browser PDF Generation**: Generate PDFs using browser's print functionality
- **Smart Experience Calculation**: Automatically calculates total work experience from individual job periods
- **Clean Design**: Minimal, professional styling optimized for both screen and print

## Files

All resume components in a single directory:

```bash
resume/
├── Makefile       # Build commands and automation
├── README.md      # This documentation
├── resume.html    # Main resume with embedded JavaScript
└── style.css      # Responsive styling and print optimization
```

## Open Source Activity

- [My issues](https://github.com/issues/created?q=is%3Aissue%20author%3A%40me%20sort%3Aupdated-desc)
- [My pull requests](https://github.com/pulls?q=is%3Apr+author%3Ayounsl+archived%3Afalse+)
