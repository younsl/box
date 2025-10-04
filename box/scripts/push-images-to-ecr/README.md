# push-images-to-ecr

Multi-architecture container image push to Amazon ECR using Podman.

## Overview

Pulls public container images (MySQL, Valkey) and pushes them to private ECR with multi-arch support (amd64/arm64) for synchronization purposes.

## Prerequisites

- AWS CLI configured with credentials
- Podman installed (recommended over Docker for better security and rootless operation)
- ECR repositories created

## Usage

```bash
export AWS_PROFILE=<YOUR_PROFILE>
./push-images-to-ecr.sh
```

## Images

- **MySQL**: `mysql:8.0.39` → ECR
- **Valkey**: `valkey/valkey:8.1.0-alpine` → ECR

## Configuration

Edit script variables:

- `AWS_REGION`: Target AWS region (default: ap-northeast-2)
- `MYSQL_SOURCE`: Source MySQL image
- `VALKEY_SOURCE`: Source Valkey image

## FAQ

### Why Podman?

This script uses Podman instead of Docker for:

- Rootless container operations
- No daemon requirement
- Better security model
- Native multi-arch manifest support
