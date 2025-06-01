# kubernetes

This directory contains [Kubernetes](https://kubernetes.io/) related resources including CLI tools, YAML manifests, helm charts, and controller source code.

## List of Contents

| Category | Name | Status | Description |
|----------|------|--------|-------------|
| Kubernetes Controller | [eip-rotation-handler](./eip-rotation-handler/) | Maintained | Kubernetes DaemonSet for rotating Public Elastic IP address of EKS worker nodes located in Public Subnet. |
| CLI Tool | [jdk-version-scanner](./jdk-version-scanner/) | Maintained | CLI tool that checks the JDK version of the running Java application pods |
| YAML Manifests | [policies](./policies/) | Maintained | Collection of [Kyverno](https://kyverno.io/) policies for Kubernetes cluster security and governance. |

## Module Status

| Category | Module | Version | Build Status |
|----------|--------|---------|--------------|
| Kubernetes Controller | [eip-rotation-handler](./kubernetes/eip-rotation-handler/) | [![GitHub release (latest by date filtered by eip-rotation-handler)](https://img.shields.io/github/v/release/younsl/box?label=eip-rotation-handler&include_prereleases&sort=semver&filter=eip-rotation-handler%2F*&style=flat-square&color=black&logo=github)](https://github.com/younsl/box/releases?q=eip-rotation-handler) | [![release-eip-rotation-handler](https://img.shields.io/github/actions/workflow/status/younsl/box/release-eip-rotation-handler.yml?label=release&style=flat-square&color=black&logo=github)](https://github.com/younsl/box/actions/workflows/release-eip-rotation-handler.yml) |