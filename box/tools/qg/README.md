# qg

A simple QR code generator that creates a QR code from a given URL.

## Usage

Build `qg` binary to use it.

```bash
go build -o qg main.go
```

Run `qg` with the URL you want to generate a QR code for.

```bash
./qg -h # or ./qg --help
./qg [flags] <url>
```