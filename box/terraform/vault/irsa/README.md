# Vault Auto Unseal with KMS

This Terraform code is designed to set up the necessary resources for Vault's Auto Unseal feature using AWS Key Management Service (KMS). It creates the following components:

- **IAM Role for Service Account (IRSA)**: This role allows the Vault service to interact with AWS services securely.
- **IAM Policy**: Defines the permissions required for the Vault service to perform operations like KMS encryption and decryption.
- **Customer-managed KMS Key**: A KMS key that Vault will use for auto-unsealing.

## Reference

This Terraform code is referenced in the blog post [Vault Auto Unseal with KMS on EKS](https://younsl.github.io/blog/vault-eks/), which provides a detailed guide on how to implement Vault's Auto Unseal feature in an EKS (Elastic Kubernetes Service) environment.

## Usage

To use this Terraform code, ensure you have the following prerequisites:

1. AWS account with necessary permissions.
2. Terraform installed on your local machine.
3. Configure AWS credentials.

Run the following commands to apply the Terraform configuration:

```bash
terraform init
terraform apply -auto-approve
```

This will provision the required resources in your AWS account.
