#===============================================================================
# External Data (Account ID, OIDC Provider ARN)
#===============================================================================
data "aws_caller_identity" "current" {}

data "aws_eks_cluster" "this" {
  name = "<CHANGE_YOUR_CLUSTER_NAME_HERE>"
}

data "aws_iam_openid_connect_provider" "this" {
  url = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

#===============================================================================
# IAM Role for Service Account (IRSA)
#===============================================================================
module "irsa_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "6.2.3"

  role_name = "vault-irsa-role"

  role_policy_arns = {
    policy = module.iam_policy.arn
  }

  oidc_providers = {
    vault = {
      provider_arn               = data.aws_iam_openid_connect_provider.this.arn
      namespace_service_accounts = ["vault:vault"]
    }
  }
}

#===============================================================================
# IAM Policy
#===============================================================================
module "iam_policy" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-policy"
  version = "6.2.3"

  name        = "vault-auto-unseal-policy"
  path        = "/"
  description = "vault pod in vault namespace to auto unseal"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "VaultAutoUnsealUsingKMS"
        Effect = "Allow"
        Action = [
          "s3:*",
          "ec2:DescribeInstances",
          "iam:*",
          "sts:*",
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })

}

#===============================================================================
# KMS Key
#===============================================================================
resource "aws_kms_key" "vault_auto_unseal" {
  description             = "KMS key for auto-unsealing Vault"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Vault to use the key"
        Effect = "Allow"
        Principal = {
          AWS = module.irsa_role.iam_role_arn
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_kms_alias" "vault_auto_unseal" {
  name          = "alias/vault-auto-unseal"
  target_key_id = aws_kms_key.vault_auto_unseal.key_id
}