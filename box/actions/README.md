# actions

## set-locale

An Actions Workflow to set locale in runner environment.

## sync-ghe-to-codecommit

An Actions Workflow to synchronize a repository located on GitHub Cloud or GitHub Enterprise Server to AWS CodeCommit.

### System Architecture

```mermaid
graph LR
    subgraph Kubernetes Cluster
        B[👷‍♂️ Actions Runner Pod]:::navy
    end
    B -->|1. clone| A[GitHub Repository]
    B -->|2. mirror push| C[CodeCommit Repository]

    classDef navy fill:#000080, color:#FFFFFF;
```
