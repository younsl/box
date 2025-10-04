# tflint

### Summary

`tflint` is a tool to check Terraform code. You can check Terraform code when you push PR by using tflint.

### Verified environment

- **Platform**: GitHub Enterprise Server (Self-hosted)
- **Actions Runner**: Self-hosted (EKS)

## Usage

1. Create a `tflint.yml` file in the `.github/workflows` directory of the Terraform repository.
2. Create a `.tflint.hcl` file in the directory containing the Terraform code. If there are multiple directories, you need to create a `.tflint.hcl` file in each directory.

```tree
.
├── my-eks-cluster
|   ├── .tflint.hcl
|   ├── main.tf
|   ├── variables.tf
|   ├── version.tf
|   └── outputs.tf
└── ... other directories ...
```

3. Write Terraform code and push PR.
4. When you push PR, tflint check will be automatically executed.
5. The result of the check will be displayed in the Files changed tab of the PR. tflint displays the result of the check by default using Problem Matcher.

> [!NOTE]
> [Problem Matcher](https://github.com/actions/toolkit/blob/main/docs/problem-matchers.md) is a feature that scans the output of the action and finds a specific pattern, and displays it prominently in the UI. When it finds matching content, it will be displayed in the GitHub comment and log file.

## References

- [tflint](https://github.com/terraform-linters/tflint)
