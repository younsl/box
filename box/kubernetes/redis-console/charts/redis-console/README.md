# redis-console

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

A Helm chart for Redis console container to manage multiple Redis clusters

**Homepage:** <https://github.com/younsl/o>

## Installation

### List available versions

This chart is distributed via OCI registry, so you need to use [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md) instead of `helm search repo` to discover available versions:

```console
crane ls ghcr.io/younsl/charts/redis-console
```

If you need to install crane on macOS, you can easily install it using [Homebrew](https://brew.sh/), the package manager.

```bash
brew install crane
```

### Install the chart

Install the chart with the release name `redis-console`:

```console
helm install redis-console oci://ghcr.io/younsl/charts/redis-console
```

Install with custom values:

```console
helm install redis-console oci://ghcr.io/younsl/charts/redis-console -f values.yaml
```

Install a specific version:

```console
helm install redis-console oci://ghcr.io/younsl/charts/redis-console --version 0.1.0
```

### Install from local chart

Download redis-console chart and install from local directory:

```console
helm pull oci://ghcr.io/younsl/charts/redis-console --untar --version 0.1.0
helm install redis-console ./redis-console
```

The `--untar` option downloads and unpacks the chart files into a directory for easy viewing and editing.

## Upgrade

```console
helm upgrade redis-console oci://ghcr.io/younsl/charts/redis-console
```

## Uninstall

```console
helm uninstall redis-console
```

## Configuration

The following table lists the configurable parameters and their default values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.registry | string | `"ghcr.io"` | Container image registry |
| image.repository | string | `"younsl/redis-console"` | Container image repository |
| image.tag | string | `""` | Container image tag (overrides the image tag whose default is the chart appVersion) |
| image.pullPolicy | string | `"Always"` | Container image pull policy |
| imagePullSecrets | list | `[]` | Image pull secrets for private container registries |
| deployment | object | `{"replicas":1}` | Deployment configuration |
| deployment.replicas | int | `1` | Number of replicas |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `"redis-console"` | The name of the service account to use |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account (e.g., IRSA for AWS permissions) |
| serviceAccount.automountServiceAccountToken | bool | `true` | Automatically mount service account token in pods |
| serviceAccount.imagePullSecrets | list | `[]` | Image pull secrets to attach to the service account |
| config.create | bool | `true` | Create Secret for cluster configuration |
| config.existingSecretName | string | `""` | Use existing Secret instead of creating a new one If set, config.create is ignored and this Secret will be used The Secret must contain a key named "config.yaml" |
| config.clusters | list | See example below | Redis cluster configurations Creates a Secret containing config.yaml at /etc/redis/clusters/config.yaml |
| config.awsRegion | string | `""` | AWS region for ElastiCache operations (optional) |
| resources | object | See below | Resource requests and limits |
| resources.requests | object | `{"cpu":"20m","memory":"30Mi"}` | Resource requests |
| resources.requests.memory | string | `"30Mi"` | Memory request |
| resources.requests.cpu | string | `"20m"` | CPU request |
| resources.limits | object | `{"memory":"60Mi"}` | Resource limits |
| resources.limits.memory | string | `"60Mi"` | Memory limit |
| securityContext | object | See below | Container-level security context configuration This security context applies to the container, not the pod |
| securityContext.allowPrivilegeEscalation | bool | `false` | Whether a process can gain more privileges than its parent process |
| securityContext.runAsNonRoot | bool | `true` | Run container as non-root user |
| securityContext.runAsUser | int | `1000` | User ID to run the container (matches Dockerfile USER directive) |
| securityContext.runAsGroup | int | `1000` | Group ID to run the container (matches Dockerfile group) |
| securityContext.capabilities | object | `{"drop":["ALL"]}` | Linux capabilities to drop |
| securityContext.capabilities.drop | list | `["ALL"]` | Drop all capabilities |
| env | object | `{}` | Environment variables |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| tolerations | list | `[]` | Tolerations for pod assignment |
| affinity | object | `{}` | Affinity rules for pod assignment |
| dnsConfig | object | `{}` | DNS configuration for the pod |
| podAnnotations | object | `{}` | Pod annotations |
| podLabels | object | `{}` | Pod labels |

## Source Code

* <https://github.com/younsl/o/edit/main/box/kubernetes/redis-console>

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
