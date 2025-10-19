# o

A monorepo containing Kubernetes tools, container images, and DevOps automation resources.

## Announcement

Migrating all Go-based CLI tools and Kubernetes controllers in this repository to [Rust](https://github.com/rust-lang/rust) 1.90. All future tools will be developed in Rust instead of Go.

**Why Rust?** Better performance and lower memory footprint, memory safety without garbage collection, superior error handling and type system, and modern tooling ecosystem.

**Migration Status**:

| Status | Applications |
|--------|--------------|
| Completed | [`kk`](./box/tools/kk), [`qg`](./box/tools/qg), [`jvs`](./box/kubernetes/jvs) (container), [`promdrop`](./box/kubernetes/promdrop) (container), [`filesystem-cleaner`](./box/containers/filesystem-cleaner) (container) |
| In Progress | [`cocd`](./box/tools/cocd), [`idled`](./box/tools/idled) |

This is a breaking change effort aimed at building a more robust and maintainable toolset.

## Featured content

Kubernetes utilities, container images, automation scripts, infrastructure code, and engineering documentation.

- **[tools](./box/tools/)** - CLI utilities (Go: cocd, idled | Rust: kk, qg)
- **[kubernetes](./box/kubernetes/)** - K8s resources, policies, and controllers (jvs, promdrop)
- **[containers](./box/containers/)** - Custom container images
- **[terraform](./box/terraform/)** - Infrastructure as Code
- **[actions](./box/actions/)** - Reusable GitHub Actions workflows
- **[scripts](./box/scripts/)** - Automation scripts for AWS, GitHub, and K8s
- **[til](./box/til/)** - Engineering notes and learnings
