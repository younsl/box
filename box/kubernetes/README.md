# kubernetes

This directory contains [Kubernetes](https://kubernetes.io/) related resources including CLI tools, YAML manifests, helm charts, and controller source code.

## List of Contents

| Category | Name | Status | Description |
|----------|------|--------|-------------|
| Kubernetes Controller | [eip-rotation-handler](./eip-rotation-handler/) | Maintained | Kubernetes DaemonSet for rotating Public Elastic IP address of EKS worker nodes located in Public Subnet. |
| CLI Tool | [jdk-version-scanner](./jdk-version-scanner/) | Maintained | CLI tool that checks the JDK version of the running Java application pods |
| YAML Manifests | [policies](./policies/) | Maintained | Collection of [Kyverno](https://kyverno.io/) policies for Kubernetes cluster security and governance. |