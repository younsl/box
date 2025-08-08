# box

Clutter box. This repository is structured as a [monorepo](https://en.wikipedia.org/wiki/Monorepo). It contains some scripts, kubernetes snippets, engineering notes, and some assets from [tech blog](https://younsl.github.io).

## Main Projects

- **Tools:**
  - [**cocd**](./box/tools/cocd): TUI for monitoring GitHub Actions deployment approvals
  - [**idled**](./box/tools/idled): CLI tool for finding idle AWS resources across regions
- **Kubernetes Addons:**
  - [**eip-rotation-handler**](./box/kubernetes/eip-rotation-handler): Kubernetes DaemonSet for AWS Elastic IP rotation
