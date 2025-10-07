# kk

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go&color=black&logoColor=white)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&color=black)](https://opensource.org/licenses/MIT)

<img src="./docs/1.png" alt="Gopher character" width="200">

kk stands for knock-knock.

kk is a CLI tool that checks the status of domains specified in a YAML configuration file. You use the `--config` flag to provide the path to your domain list for validation.

## Standard Layout

This project follows the [Golang Standard Project Layout](https://github.com/golang-standards/project-layout) to keep best practices.

```console
.
├── cmd/
│   └── kk/
├── internal/
│   ├── checker/
│   └── config/
├── configs/
|   └── domain-example.yaml
└── docs/
```

## Usage

Build CLI tool locally:

```bash
make build
```

Run CLI tool:

```console
$ ./kk --config configs/domain-example.yaml
Loaded domain list from 'configs/domain-example.yaml'.
URL                            TIME   STATUS           CODE  ATTEMPTS
https://www.github.com         205ms  OK               200   1
https://registry.k8s.io/v2/    237ms  OK               200   1
https://www.google.com         401ms  OK               200   1
https://stackoverflow.com      591ms  OK               200   1
https://www.stackoverflow.com  830ms  OK               200   1
https://reddit.com             324ms  UNEXPECTED_CODE  403   3 (failed)

Summary: 5/6 successful checks in 5.3s.
```

You can specify the path to your configuration file using the `--config` flag.

## Configuration

The configuration file should be in YAML format and contain a list of domains under the `domains` key. See [configs/domain-example.yaml](./configs/domain-example.yaml) for an example.

> [!NOTE]
> If you omit `http://` or `https://` in the domain, kk CLI will automatically add `https://` to the domain.

```yaml
# configs/domain-example.yaml
domains:
  # If you omit http:// or https://, kk will automatically add https:// to the domain.
  - www.google.com
  - reddit.com
  - https://registry.k8s.io/v2/
```