# actions

## set-locale

> It is designed to run in a **GitHub Enterprise Server** environment.

An Actions Workflow to set locale in runner environment.

## sync-ghe-to-codecommit

> It is designed to run in a **GitHub Enterprise Server** environment.

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

## terraform-apply-cron

> It is designed to run in a **GitHub Enterprise Server** environment.

This GitHub Actions workflow is set up to automatically run a Terraform script every day at 1 AM KST.

```yaml
on:
  schedule:
    # Run KST 01:00 AM by cron trigger
    - cron:  '0 16 * * *'
```

It starts by checking out the code, setting up node.js and terraform, and configuring AWS credentials. Then, it formats<sup>`fmt`</sup> and `validate` the terraform code, and `apply` terraform codes located in the specified path.

```mermaid
graph LR
    subgraph Kubernetes Cluster
        A[👷‍♂️ Actions Runner Pod]:::navy
    end
    T[Time] -->|⏱️ Cron trigger| A
    A -->|1. clone terraform code| B[GitHub Repository]
    A -->|2. Get permissions| C[IAM Key]
    C -->|3. terraform init, apply| D[AWS Resources]

    classDef navy fill:#000080, color:#FFFFFF;
```
