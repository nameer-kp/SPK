# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

## [0.2.0] - 2025-02-04

### Added
- Project discovery via `FLOW_PROJECT_DIRS` environment variable
- Environment selection screen in TUI
- List selection screen for workflow select blocks
- Example project configs and workflows

### Changed
- Improved project and environment selection flow in TUI

## [0.1.0] - 2025-01-XX

### Added
- Initial release
- Core workflow engine with YAML parsing
- Template variable resolution with `{{}}` syntax
- Built-in node types:
  - `http` - HTTP requests with retry support
  - `db` - Database queries (PostgreSQL, MySQL, SQLite)
  - `shell` - Shell command execution
  - `file` - File operations (read, write, append, exists)
  - `delay` - Workflow pause/delay
  - `transform` - Data transformation via expressions
  - `command` - Multi-project command execution
- Interactive TUI with Bubble Tea
- Multi-level configuration system
- Profile support for environment-specific configs
- Node registry with plugin-like architecture

[Unreleased]: https://github.com/nameer-kp/atem/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/nameer-kp/atem/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/nameer-kp/atem/releases/tag/v0.1.0
