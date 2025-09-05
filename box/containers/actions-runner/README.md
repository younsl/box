# GitHub Actions Runner with Additional APT Sources

Custom [GitHub Actions runner image](https://github.com/actions/actions-runner-controller/tree/master/runner) with additional APT mirror sources for improved package availability and download speeds.

## Base Image

- `summerwind/actions-runner:v2.328.0-ubuntu-22.04`

## Features

- Adds Kakao mirror for faster package downloads in Korea
- Maintains compatibility with the original runner image
- Easy to extend with more mirror sources

## Build

```bash
docker build -t actions-runner-kakao . --platform linux/amd64
```

## Usage

This image can be used as a drop-in replacement for the standard GitHub Actions runner image in self-hosted runner deployments.

## Customization

To add or modify mirror sources, edit the `additional-sources.list` file.

## References

- [GitHub Community Discussion on APT mirrors](https://github.com/orgs/community/discussions/160684)
