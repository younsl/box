# resume

My minimalistic resume built with native HTML and CSS.

## Usage

Generate resume pdf file from resume html file by running the following command:

```bash
make help
make pdf
```

## Troubleshooting

### Disable header and footer in resume pdf file

If you faced the issue printing footer and header in resume pdf file, you can try to run the chrome browser with the `--print-to-pdf-no-header` flag.

```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --headless --disable-gpu --print-to-pdf-no-header --print-to-pdf=$(OUTPUT_PDF) $(RESUME_HTML)
```

In headless chrome browser, `--print-to-pdf-no-header` flag is required to print the resume page without the header and footer.