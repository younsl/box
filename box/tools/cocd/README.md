# cocd

[![Go Version](https://img.shields.io/badge/go-1.24-000000?style=flat-square&logo=go&logoColor=white)](go.mod)
[![GitHub release](https://img.shields.io/github/v/release/younsl/box?style=flat-square&color=black&logo=github&logoColor=white&label=release)](https://github.com/younsl/box/releases?q=cocd)
[![License](https://img.shields.io/github/license/younsl/box?style=flat-square&color=black&logo=github&logoColor=white)](/LICENSE)

> cocd stands for Chaos Of Continuous Deployment

A TUI (Terminal User Interface) application for monitoring GitHub Actions jobs that are waiting for approval in production environments. Inspired by [k9s](https://github.com/derailed/k9s), cocd provides an interactive interface to monitor your GitHub Actions workflows in real-time.

## Background

GitHub Actions runs separately in each repository, unlike Jenkins which shows everything in one place. This makes it hard for DevOps and SRE teams to see what's happening across all their projects, especially in companies where a central team approves and manages all production deployments.

cocd was built to solve this problem. Just like k9s helps you manage Kubernetes clusters from your terminal, cocd lets you monitor and control GitHub Actions deployments across all your repositories from one terminal window.

## Features

DevOps Engineers and SREs can use cocd to manage GitHub Actions workflows through a simple terminal interface:

- **Approval waiting job monitoring** - Monitor GitHub Actions jobs waiting for approval
- **Recent Actions job monitoring** - View recent workflow runs and their status
- **Job approval** - Approve pending deployment jobs directly from the TUI
- **Job cancellation** - Cancel running or pending jobs
- **Real-time updates** - Live monitoring with configurable refresh intervals
- **Multi-environment support** - Monitor different environments (prod, staging, dev)

## Architecture

<img width="676" height="265" alt="image" src="https://github.com/user-attachments/assets/003b6092-f25a-4672-b10d-0b7526cae163" />

## Documentation

Comprehensive guides and references for using cocd effectively.

- [Configuration](docs/configuration.md): Setup and configuration guide
- [Roadmap](docs/roadmap.md): Future development plans and upcoming features 
