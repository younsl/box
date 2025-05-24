# tools

Kubernetes and automation CLI tools.

## Tool List

| Category | Name | Description |
|----------|------|-------------|
| Kubernetes | [eip-rotation-handler](./eip-rotation-handler/) | ERH(EIP Rotation Handler) is a daemonset that handles Elastic IP rotation for Kubernetes public nodes to avoid IP-based rate limits for 3rd party services. |
| Kubernetes | [jdk-version-scanner](./jdk-version-scanner/) | Scans pods in specified Kubernetes namespaces for installed JDK versions. Uses `kubectl exec` to run `java -version`, filters out DaemonSet pods, and outputs a summary table. |
| CLI | [kk](./kk/) (knock knock) | CLI tool that checks the status of domains specified in a YAML configuration file. |
| CLI | [qg](./qg/) (qr generator) | CLI tool that generates QR code images from text or URLs. |
