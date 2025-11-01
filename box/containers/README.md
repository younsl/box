# Container Images

This directory contains custom container images for various DevOps and development purposes.

Inspired by [bitnami/containers](https://github.com/bitnami/containers).

## Available Images

Production-ready container images for DevOps automation, development tooling, and Kubernetes workloads. Published to [public ghcr.io](https://github.com/younsl?tab=packages) (GitHub Container Registry) or stored locally.

| # | Image Name | Description | Helm Chart | Remark |
|---|------------|-------------|------------|--------|
| 1 | [**actions-runner**](./actions-runner/) | Custom actions-runner with additional tools | [actions-runner](https://github.com/younsl/charts/tree/main/charts/actions-runner) | [ghcr.io/younsl/actions-runner](https://github.com/younsl/o/pkgs/container/actions-runner) |
| 2 | [**ab**](./ab/) | Lightweight load testing | - | - |
| 3 | [**mageai**](./mageai/0.9.73-custom.1/) | Custom mageai 0.9.73 image | - | - |
| 4 | [**backup-utils**](/.github/workflows/release-backup-utils.yml) | GitHub Enterprise backup/restore tools (uses original Dockerfile) | [backup-utils](https://github.com/younsl/charts/tree/main/charts/backup-utils) | [ghcr.io/younsl/backup-utils](https://github.com/younsl/o/pkgs/container/backup-utils) ⚠️ **Deprecated** - GitHub Enterprise Server 3.17+ includes [built-in backup service](https://docs.github.com/en/enterprise-server@latest/admin/backing-up-and-restoring-your-instance/backup-service-for-github-enterprise-server/configuring-the-backup-service) |
| 5 | [**filesystem-cleaner**](./filesystem-cleaner/) | Sidecar container that monitors and cleans specified directories | - | [ghcr.io/younsl/filesystem-cleaner](https://github.com/younsl/o/pkgs/container/filesystem-cleaner) |

## References

- **Helm Charts**: [younsl/charts](https://github.com/younsl/charts) - Helm charts repository maintained by me (younsl) that uses these container images (actions-runner, backup-utils)

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file in the project root for details.
