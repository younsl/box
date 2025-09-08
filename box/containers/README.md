# Container Images

![logo](https://github.com/younsl/younsl.github.io/blob/main/content/slides/admission-policy/assets/pink-container-84x84.png)

This directory contains custom container images for various DevOps and development purposes.

## Available Images

Custom container images built for specific DevOps workflows and development needs. Images are either stored locally or pushed to [public ghcr.io](https://github.com/younsl?tab=packages) (GitHub Container Registry).

| # | Image | Purpose | Base Image | Description | Remark |
|---|-------|---------|------------|-------------|--------|
| 1 | [**actions-runner**](./actions-runner/) | GitHub Actions self-hosted runner | [summerwind/actions-runner](https://github.com/actions/actions-runner-controller/tree/master/runner) | Custom runner with additional tools | [ghcr.io/younsl/actions-runner](https://github.com/younsl/box/pkgs/container/actions-runner), [helm chart](https://github.com/younsl/charts/tree/main/charts/actions-runner) |
| 2 | [**ab**](./ab/) | Apache Bench load testing tool | Alpine | Lightweight load testing | - |
| 3 | [**hugo**](./hugo/) | Hugo static site generator | Alpine | Fast static site builds | - |
| 4 | [**mageai**](./mageai/0.9.73-custom.1/) | Mage AI data pipeline platform | mageai/mageai:0.9.73 | Custom version 0.9.73 | - |
| 5 | [**terraform-console-machine**](./terraform-console-machine/) | Terraform development environment | hashicorp/terraform | Interactive Terraform console | - |
| 6 | [**yarn**](./yarn/) | Yarn package manager | node | Node.js with Yarn | - |
| 7 | [**backup-utils**](/.github/workflows/release-backup-utils.yml) | GitHub Enterprise backup utilities | [github/backup-utils](https://github.com/github/backup-utils/releases) (unmodified) | GitHub Enterprise backup/restore tools (uses original Dockerfile) | [ghcr.io/younsl/backup-utils](https://github.com/younsl/box/pkgs/container/backup-utils), [helm chart](https://github.com/younsl/charts/tree/main/charts/backup-utils) |
| 8 | [**filesystem-cleaner**](./filesystem-cleaner/) | Temporary file cleanup for Kubernetes | golang:1.25-alpine | Sidecar container that monitors and cleans specified directories | - |

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file in the project root for details.
