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

---

## Open Source Contributions

### Activity

- [My open pull requests](https://github.com/pulls?q=is%3Aopen+is%3Apr+author%3Ayounsl+archived%3Afalse+-org%3Ayounsl)
- [My merged pull requests](https://github.com/pulls?q=is%3Apr+author%3Ayounsl+archived%3Afalse+-org%3Ayounsl+-org%3Acontainerelic+is%3Amerged+)

### Notable Projects:

- [charts](https://younsl.github.io/charts): Helm charts
- [gss](https://github.com/containerelic/gss): GHES Schedule Scanner for scanning scheduled workflow in Github Enterprise Server
- [eip-rotation-handler](https://github.com/younsl/box/tree/main/box/kubernetes/eip-rotation-handler): Kubernetes DaemonSet for rotating Public Elastic IP address of EKS worker nodes located in Public Subnet
- [blog](https://younsl.github.io/): Tech blog focusing on Kubernetes and AWS