# kubernetes

<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/kubernetes/kubernetes-plain.svg" width="40" height="40"/>

This directory contains [Kubernetes](https://kubernetes.io/) related resources including CLI tools, YAML manifests, helm charts, and controller source code.

## List of Contents

Kubernetes tools, policy resources, and architecture documentation organized by category.

| Category | Name | Language | Status | Description |
|----------|------|----------|--------|-------------|
| Kubernetes Addon | [ec2-statuscheck-rebooter](./ec2-statuscheck-rebooter/) | [Rust](./ec2-statuscheck-rebooter/Cargo.toml) | Active | Automated reboot for standalone EC2 instances with status check failures running as Kubernetes [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) |
| Kubernetes Addon | [elasticache-backup](./elasticache-backup/) | [Rust](./elasticache-backup/Cargo.toml) | Active | ElastiCache snapshot backup to S3 automation running as Kubernetes [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) |
| Tools | [podver](./podver/) | [Rust](./podver/Cargo.toml) | Active | CLI tool that scans and reports Java and Node.js runtime versions across pods in a cluster |
| Tools | [promdrop](./promdrop/) | [Rust](./promdrop/Cargo.toml) | Active | Prometheus scrape config generator to drop unused metrics analyzed by [mimirtool](https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/) |
| Resources | [policies](./policies/) | - | Active | Collection of [Kyverno](https://kyverno.io/) policies for Kubernetes cluster security and governance. |
| Documentation | [mermaids](./mermaids/) | - | Active | Mermaid diagrams explaining Kubernetes component relationships and architecture. |

## License

All tools and resources in this directory are licensed under the repository's main [MIT License](../../LICENSE).
