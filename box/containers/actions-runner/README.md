# Actions Runner

[![GHCR](https://img.shields.io/badge/ghcr.io-younsl%2Factions--runner-000000?style=flat-square&logo=github&logoColor=white)](https://github.com/younsl/o/pkgs/container/actions-runner)
[![License](https://img.shields.io/github/license/younsl/o?style=flat-square&color=black)](https://github.com/younsl/o/blob/main/LICENSE)

GitHub Actions runner optimized for Korea with multiple APT package sources for faster and more reliable downloads.

## What's Inside

**Base:** `summerwind/actions-runner:v2.329.0-ubuntu-24.04`

**Additions:**
- Multiple APT package sources (includes Kakao mirror)
- Build essentials (`make`)

## Why This Image

- **Faster downloads** in Korea with local package sources
- **High availability** through multiple package servers
- **Drop-in replacement** for standard runner images

## Quick Start

```bash
# Build
docker build -t actions-runner . --platform linux/amd64

# Use in your runner deployment
# (Replace summerwind/actions-runner with this image)
```

## Customization

Edit `additional-sources.list` to add or modify APT repository sources.

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for version history and release notes.

## References

- [Actions Runner Available Images](https://github.com/actions/runner-images?tab=readme-ov-file#available-images)
