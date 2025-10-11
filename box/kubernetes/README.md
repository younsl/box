# kubernetes

<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/kubernetes/kubernetes-plain.svg" width="40" height="40"/>

This directory contains [Kubernetes](https://kubernetes.io/) related resources including CLI tools, YAML manifests, helm charts, and controller source code.

## List of Contents

Kubernetes tools, policy resources, and architecture documentation organized by category.

| Category | Name | Status | Description |
|----------|------|--------|-------------|
| Tools | [promdrop](./promdrop/) | Active | Prometheus scrape config generator to drop unused metrics analyzed by [mimirtool](https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/) |
| Resources | [policies](./policies/) | Active | Collection of [Kyverno](https://kyverno.io/) policies for Kubernetes cluster security and governance. |
| Tools | [jdk-version-scanner](./jdk-version-scanner/) | Active | CLI tool that checks the JDK version of the running Java application pods |
| Documentation | [mermaids](./mermaids/) | Active | Mermaid diagrams explaining Kubernetes component relationships and architecture. |

## License

All tools and resources in this directory are licensed under the repository's main [MIT License](../../LICENSE).
