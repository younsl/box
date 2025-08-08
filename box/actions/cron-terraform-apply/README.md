## terraform-apply-cron

### Summary

This GitHub Actions workflow is set up to automatically run a Terraform script every day at 1 AM KST.

```yaml
on:
  schedule:
    # Run KST 01:00 AM by cron trigger
    - cron:  '0 16 * * *'
```

### Verified environment

- **Platform**: GitHub Enterprise Server (Self-hosted)
- **Actions Runner**: Self-hosted (EKS)