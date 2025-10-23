# o

A monorepo containing Kubernetes tools, container images, and DevOps automation resources.

## Built with Rust

All applications in [`kubernetes/`](./box/kubernetes/), [`tools/`](./box/tools/), and [`containers/`](./box/containers/) are built with **[Rust](https://github.com/rust-lang/rust) 1.90+** (except `cocd` which uses Go).

Rust provides key operational benefits: minimal container sizes, low memory footprint, single static binaries with no runtime dependencies, memory safety preventing null pointer and buffer overflow crashes, and compile-time guarantees ensuring system stability in production.

## Featured content

Kubernetes utilities, container images, automation scripts, infrastructure code, and engineering documentation.

- **[tools](./box/tools/)** - CLI utilities (Go: cocd | Rust: kk, qg | Archived: idled)
- **[kubernetes](./box/kubernetes/)** - K8s resources, policies, and controllers (podver, promdrop, elasticache-backup, policies)
- **[containers](./box/containers/)** - Custom container images (actions-runner, filesystem-cleaner, hugo, ab, mageai, yarn, terraform-console-machine)
- **[terraform](./box/terraform/)** - Infrastructure as Code
- **[actions](./box/actions/)** - Reusable GitHub Actions workflows
- **[scripts](./box/scripts/)** - Automation scripts for AWS, GitHub, and K8s
- **[til](./box/til/)** - Engineering notes and learnings
