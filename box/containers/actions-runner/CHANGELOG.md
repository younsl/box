# Changelog

All notable changes to this project will be documented in this file.

## younsl-0.4.0

| Item | Value |
|------|-------|
| **Release Date** | 2025-11-03 KST |
| **Runner Version** | v2.329.0 |
| **Ubuntu Version** | 24.04 LTS |
| **Tag** | `actions-runner/v2.329.0-ubuntu-24.04-younsl-0.4.0` |

### Changed

- Migrated APT sources configuration from traditional format to [DEB822 format](https://repolib.readthedocs.io/en/latest/deb822-format.html) (`kakao-mirror.sources`)
- Enhanced security by explicitly specifying GPG key with `Signed-By` field
- Added explicit `Enabled` field for better configuration clarity
- Simplified Dockerfile by removing temporary file operations

## younsl-0.3.0

| Item | Value |
|------|-------|
| **Release Date** | 2025-11-01 KST |
| **Runner Version** | v2.329.0 |
| **Ubuntu Version** | 24.04 LTS |
| **Tag** | `actions-runner/v2.329.0-ubuntu-24.04-younsl-0.3.0` |

### Changed

- Upgraded base image from `summerwind/actions-runner:v2.328.0-ubuntu-22.04` to `summerwind/actions-runner:v2.329.0-ubuntu-24.04`
- Upgraded Actions Runner version from v2.328.0 to v2.329.0
- Upgraded Ubuntu from 22.04 LTS to 24.04 LTS for version lifecycle support

## younsl-0.2.0

| Item | Value |
|------|-------|
| **Release Date** | 2025-09-06 KST |
| **Runner Version** | v2.328.0 |
| **Ubuntu Version** | 22.04 LTS |
| **Tag** | `actions-runner/v2.328.0-ubuntu-22.04-younsl-0.2.0` |

### Features

- Additional APT sources for better package availability [#160684](https://github.com/orgs/community/discussions/160684)
- Pre-installed `make` and build essentials
- Custom GitHub Actions runner based on [summerwind/actions-runner](https://hub.docker.com/r/summerwind/actions-runner/tags)
