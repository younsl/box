# cocd Development Roadmap

## Planned Features

Not yet

## Completed Features

### v0.3.0 - Minor Feature Release

**Completed**: Auto-generated Configuration (CRF-1) + Real-time Performance Enhancement (CRF-2)

**CRF-1 Features Delivered**:
- Automatic creation of `$HOME/.config/cocd/config.yaml` on first run
- Config path resolution with priority order
- Default config template generation
- Directory creation with proper permissions
- Skeleton config with sensible defaults

**CRF-2 Features Delivered**:
- Real-time monitoring with repository scope
- Optimized scanning performance for immediate feel
- Enhanced update mechanisms for faster response
- Improved UI responsiveness during monitoring
- Repository-focused workflow scanning

**Implementation**: See `internal/config/config.go` and `internal/config/skeleton.go` for config features

### v0.2.0 - Intermediate Release

**Completed**: Foundational improvements and stability fixes

### v0.1.0 - Core TUI Framework

**Completed**: Basic TUI functionality

**Features Delivered**:
- Terminal User Interface for monitoring GitHub Actions
- Real-time workflow run monitoring
- Job approval and cancellation capabilities
- Multi-environment support
- Configurable refresh intervals

## Release Timeline

| Version | Target | Status | Features |
|---------|--------|--------|---------|
| v0.1.0  | Released | ✅ Completed | Core TUI, basic monitoring |
| v0.2.0  | Released | ✅ Completed | Stability improvements |
| v0.3.0  | Released | ✅ Completed | Auto-config + Real-time enhancements |
