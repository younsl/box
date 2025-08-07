# cocd Development Roadmap

## Planned Features

### CRF-1. Auto-generate Default Configuration

**Target**: v0.3.0

Automatically create `$HOME/.config/cocd/config.yaml` on first run cocd command.

**Tasks**
- Implement config path resolution
- Create default config template
- Add directory creation with proper permissions
- Implement interactive setup for missing values

### CRF-2. Real-time Scan Performance Enhancement

**Target**: v0.3.0

Improve scanning performance for real-time monitoring feel.

**Tasks**
- Profile performance bottlenecks
- Implement parallel workflow fetching
- Add incremental update mechanism
- Create visual scanning indicators
- Implement intelligent caching layer

**Performance Targets**
- Reduce full scan time by 50%
- Update latency < 2 seconds
- Maintain smooth UI rendering

## Completed Features

TBA

## Release Timeline

| Version | Target | Status |
|---------|--------|--------|
| v0.3.0  | Q3 2025 | In Development |
