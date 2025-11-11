# ec2-statuscheck-rebooter

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Automated reboot for standalone EC2 instances outside Kubernetes cluster

**Homepage:** <https://github.com/younsl/o>

## Installation

### List available versions

This chart is distributed via OCI registry, so you need to use [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md) instead of `helm search repo` to discover available versions:

```console
crane ls ghcr.io/younsl/charts/ec2-statuscheck-rebooter
```

If you need to install crane on macOS, you can easily install it using [Homebrew](https://brew.sh/), the package manager.

```bash
brew install crane
```

### Install the chart

Install the chart with the release name `ec2-statuscheck-rebooter`:

```console
helm install ec2-statuscheck-rebooter oci://ghcr.io/younsl/charts/ec2-statuscheck-rebooter
```

Install with custom values:

```console
helm install ec2-statuscheck-rebooter oci://ghcr.io/younsl/charts/ec2-statuscheck-rebooter -f values.yaml
```

Install a specific version:

```console
helm install ec2-statuscheck-rebooter oci://ghcr.io/younsl/charts/ec2-statuscheck-rebooter --version 0.1.0
```

### Install from local chart

Download ec2-statuscheck-rebooter chart and install from local directory:

```console
helm pull oci://ghcr.io/younsl/charts/ec2-statuscheck-rebooter --untar --version 0.1.0
helm install ec2-statuscheck-rebooter ./ec2-statuscheck-rebooter
```

The `--untar` option downloads and unpacks the chart files into a directory for easy viewing and editing.

## Upgrade

```console
helm upgrade ec2-statuscheck-rebooter oci://ghcr.io/younsl/charts/ec2-statuscheck-rebooter
```

## Uninstall

```console
helm uninstall ec2-statuscheck-rebooter
```

## Configuration

The following table lists the configurable parameters and their default values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override ec2-statuscheck-rebooter.fullname |
| fullnameOverride | string | `""` | String to fully override ec2-statuscheck-rebooter.fullname |
| replicaCount | int | `1` | Number of replicas |
| revisionHistoryLimit | int | `10` | Number of ReplicaSets to retain for rollback |
| strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | Deployment strategy |
| image.repository | string | `"ghcr.io/younsl/ec2-statuscheck-rebooter"` | Image repository |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| imagePullSecrets | list | `[]` | Image pull secrets |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.automountServiceAccountToken | bool | `true` | Automatically mount a ServiceAccount's API credentials |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.imagePullSecrets | list | `[]` | Image pull secrets to attach to the service account |
| serviceAccount.name | string | `""` | The name of the service account to use If not set and create is true, a name is generated using the fullname template |
| podAnnotations | object | `{}` | Annotations to add to the pod |
| podLabels | object | `{}` | Labels to add to the pod |
| priorityClassName | string | `""` | Priority class name for pod scheduling |
| runtimeClassName | string | `""` | Runtime class name for the pod |
| podSecurityContext | object | `{"fsGroup":1000,"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` | Pod Security Context |
| podSecurityContext.runAsNonRoot | bool | `true` | Run container as non-root user |
| podSecurityContext.runAsUser | int | `1000` | User ID for the pod |
| podSecurityContext.runAsGroup | int | `1000` | Group ID for the pod |
| podSecurityContext.fsGroup | int | `1000` | FSGroup that owns the pod's volumes |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | Security Context for container |
| securityContext.allowPrivilegeEscalation | bool | `false` | Prevent privilege escalation |
| securityContext.capabilities | object | `{"drop":["ALL"]}` | Drop all capabilities |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount root filesystem as read-only |
| extraEnv | list | `[]` | Additional environment variables for the container |
| rebooter.checkIntervalSeconds | int | `300` | Check interval in seconds between status checks |
| rebooter.failureThreshold | int | `2` | Number of consecutive failures before reboot |
| rebooter.region | string | `""` | AWS region (leave empty to use default from IRSA/instance metadata) |
| rebooter.tagFilters | list | `[]` | Comma-separated tag filters for EC2 instances (format: Key=Value) |
| rebooter.dryRun | bool | `false` | Dry run mode (no actual reboot will be performed) |
| rebooter.logFormat | string | `"json"` | Log format: json or pretty |
| rebooter.logLevel | string | `"info"` | Log level: trace, debug, info, warn, error |
| livenessProbe.initialDelaySeconds | int | `10` | Initial delay before liveness probe starts |
| livenessProbe.periodSeconds | int | `30` | How often to perform the probe |
| livenessProbe.timeoutSeconds | int | `5` | Timeout for the probe |
| livenessProbe.failureThreshold | int | `3` | Number of failures before marking as unhealthy |
| readinessProbe.initialDelaySeconds | int | `5` | Initial delay before readiness probe starts |
| readinessProbe.periodSeconds | int | `10` | How often to perform the probe |
| readinessProbe.timeoutSeconds | int | `5` | Timeout for the probe |
| readinessProbe.failureThreshold | int | `3` | Number of failures before marking as not ready |
| resources.limits.memory | string | `"128Mi"` | Memory limit |
| resources.requests.cpu | string | `"100m"` | CPU request |
| resources.requests.memory | string | `"64Mi"` | Memory request |
| resizePolicy | list | `[]` | Container resize policy for CPU and memory Allows in-place resource updates without pod restart (beta since Kubernetes 1.33) |
| dnsPolicy | string | `""` | DNS policy for the pod Options: Default, ClusterFirst (default), ClusterFirstWithHostNet, None |
| dnsConfig | object | `{}` | DNS configuration for the pod |
| volumes | list | `[]` | Additional volumes for the pod |
| volumeMounts | list | `[]` | Additional volume mounts for the container |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| tolerations | list | `[]` | Tolerations for pod assignment |
| affinity | object | `{}` | Affinity for pod assignment |

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
