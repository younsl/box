# elasticache-backup

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

ElastiCache snapshot backup to S3 automation

**Homepage:** <https://github.com/younsl/o>

## Installation

### List available versions

This chart is distributed via OCI registry, so you need to use [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md) instead of `helm search repo` to discover available versions:

```console
crane ls ghcr.io/younsl/charts/elasticache-backup
```

If you need to install crane on macOS, you can easily install it using [Homebrew](https://brew.sh/), the package manager.

```bash
brew install crane
```

### Install the chart

Install the chart with the release name `elasticache-backup`:

```console
helm install elasticache-backup oci://ghcr.io/younsl/charts/elasticache-backup
```

Install with custom values:

```console
helm install elasticache-backup oci://ghcr.io/younsl/charts/elasticache-backup -f values.yaml
```

Install a specific version:

```console
helm install elasticache-backup oci://ghcr.io/younsl/charts/elasticache-backup --version 0.1.0
```

### Install from local chart

Download elasticache-backup chart and install from local directory:

```console
helm pull oci://ghcr.io/younsl/charts/elasticache-backup --untar --version 0.1.0
helm install elasticache-backup ./elasticache-backup
```

The `--untar` option downloads and unpacks the chart files into a directory for easy viewing and editing.

## Upgrade

```console
helm upgrade elasticache-backup oci://ghcr.io/younsl/charts/elasticache-backup
```

## Uninstall

```console
helm uninstall elasticache-backup
```

## Configuration

The following table lists the configurable parameters and their default values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image | object | `{"pullPolicy":"Always","registry":"ghcr.io","repository":"younsl/elasticache-backup","tag":""}` | Container image configuration |
| image.registry | string | `"ghcr.io"` | Container image registry |
| image.repository | string | `"younsl/elasticache-backup"` | Container image repository |
| image.tag | string | `""` | Container image tag (overrides the image tag whose default is the chart appVersion) |
| image.pullPolicy | string | `"Always"` | Container image pull policy |
| imagePullSecrets | list | `[]` | Image pull secrets for private container registries |
| elasticache | object | `{"cacheClusterId":"","region":"ap-northeast-2"}` | ElastiCache configuration |
| elasticache.cacheClusterId | string | `""` | ElastiCache cluster ID (read replica node) - REQUIRED |
| elasticache.region | string | `"ap-northeast-2"` | AWS region where ElastiCache cluster is located |
| s3 | object | `{"bucketName":""}` | S3 configuration |
| s3.bucketName | string | `""` | S3 bucket name for storing RDB files - REQUIRED |
| cronjob | object | `{"activeDeadlineSeconds":3600,"backoffLimit":2,"concurrencyPolicy":"Forbid","failedJobsHistoryLimit":3,"restartPolicy":"OnFailure","schedule":"10 15 * * *","successfulJobsHistoryLimit":3,"suspend":false,"timeZone":""}` | CronJob schedule configuration |
| cronjob.suspend | bool | `false` | Suspend CronJob execution (useful for maintenance) |
| cronjob.schedule | string | `"10 15 * * *"` | Cron schedule expression (default: daily at 00:10 KST / 15:10 UTC) |
| cronjob.timeZone | string | `""` | Timezone for schedule (requires Kubernetes 1.25+). Examples: "Asia/Seoul", "UTC", "America/New_York" |
| cronjob.successfulJobsHistoryLimit | int | `3` | Number of successful job history to retain |
| cronjob.failedJobsHistoryLimit | int | `3` | Number of failed job history to retain |
| cronjob.concurrencyPolicy | string | `"Forbid"` | Concurrency policy for CronJob (Allow, Forbid, Replace) |
| cronjob.restartPolicy | string | `"OnFailure"` | Restart policy for failed jobs |
| cronjob.backoffLimit | int | `2` | Number of retries before marking job as failed |
| cronjob.activeDeadlineSeconds | int | `3600` | Maximum duration in seconds for job to complete (1 hour) |
| snapshot | object | `{"checkInterval":30,"exportTimeout":300,"retentionCount":7,"timeout":1800}` | Snapshot operation configuration |
| snapshot.timeout | int | `1800` | Maximum wait time for snapshot completion in seconds (30 minutes) |
| snapshot.exportTimeout | int | `300` | Maximum wait time for S3 export completion in seconds (5 minutes) |
| snapshot.checkInterval | int | `30` | Snapshot status check interval in seconds |
| snapshot.retentionCount | int | `7` | Number of snapshots to retain in S3 (0 = unlimited, no cleanup) |
| serviceAccount | object | `{"annotations":{},"automountServiceAccountToken":true,"create":true,"name":"elasticache-backup"}` | Service Account configuration |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `"elasticache-backup"` | The name of the service account to use |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account (e.g., IRSA for AWS permissions) |
| serviceAccount.automountServiceAccountToken | bool | `true` | Automatically mount service account token in pods |
| resources | object | See below | Resource requests and limits |
| resources.requests | object | `{"cpu":"100m","memory":"128Mi"}` | Resource requests |
| resources.requests.memory | string | `"128Mi"` | Memory request |
| resources.requests.cpu | string | `"100m"` | CPU request |
| resources.limits | object | `{"memory":"256Mi"}` | Resource limits |
| resources.limits.memory | string | `"256Mi"` | Memory limit |
| securityContext | object | See below | Security context for the container |
| securityContext.allowPrivilegeEscalation | bool | `false` | Whether a process can gain more privileges than its parent process |
| securityContext.runAsNonRoot | bool | `true` | Run container as non-root user |
| securityContext.runAsUser | int | `1000` | User ID to run the container |
| securityContext.capabilities | object | `{"drop":["ALL"]}` | Linux capabilities to drop |
| securityContext.capabilities.drop | list | `["ALL"]` | Drop all capabilities |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount root filesystem as read-only |
| env | object | `{"logFormat":"json","logLevel":"info","timezoneOffsetHours":9}` | Environment variables |
| env.logLevel | string | `"info"` | Log level (debug, info, warn, error) |
| env.logFormat | string | `"json"` | Log format (json or pretty) |
| env.timezoneOffsetHours | int | `9` | Timezone offset in hours for snapshot filename generation (e.g., 9 for Asia/Seoul UTC+9, 0 for UTC) |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| tolerations | list | `[]` | Tolerations for pod assignment |
| affinity | object | `{}` | Affinity rules for pod assignment |
| dnsConfig | object | `{}` | DNS configuration for the pod |

## Source Code

* <https://github.com/younsl/o/tree/main/box/kubernetes/elasticache-backup>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| younsl | <cysl@kakao.com> | <https://github.com/younsl> |

## License

This chart is licensed under the Apache License 2.0. See [LICENSE](https://github.com/younsl/o/blob/main/LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a [Pull Request](https://github.com/younsl/o/pulls).

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
