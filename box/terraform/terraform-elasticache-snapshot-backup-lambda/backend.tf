terraform {
  backend "s3" {
    encrypt = true
    region  = "<YOUR_REGION>"
    bucket  = "<YOUR_BUCKET_NAME>"
    key     = "<YOUR_BUCKET_KEY_PATH>/terraform.tfstate"

    # S3 state lock feature is only available in Terraform CLI version 1.10.0 and above.
    use_lockfile = true
  }
}
