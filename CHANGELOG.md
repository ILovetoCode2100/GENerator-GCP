# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive project cleanup and reorganization
- Contributing guidelines
- MIT License
- Improved documentation structure

### Changed
- Restructured documentation into organized subdirectories
- Updated README to be more generic and professional
- Consolidated redundant documentation files

### Removed
- Client-specific content and examples
- Development and debug files
- Temporary and test-specific scripts

## [1.0.0] - 2025-01-09

### Added
- Complete API CLI implementation
- All core commands (create, list, update operations)
- Batch structure creation from YAML/JSON
- Multiple output formats (human, JSON, YAML, AI)
- Comprehensive step creation commands (40+ step types)
- Auto-attachment of checkpoints to journeys
- Journey renaming and navigation URL updates
- Docker support with Alpine-based image
- Extensive documentation and examples

### Features
- **Project Management**: Create and list projects
- **Goal Management**: Create goals with auto-journey creation
- **Journey Management**: Create, list, and update journeys
- **Checkpoint Management**: Create and list checkpoints with auto-attach
- **Step Creation**: Support for multiple step types:
  - Navigation steps
  - Click actions (single, double, right-click, hover)
  - Text input and keyboard actions
  - Wait operations (time-based and element-based)
  - Assertions (exists, equals, checked, etc.)
  - Scroll actions
  - Data operations (store values, execute JavaScript)
  - Browser management (cookies, alerts, window sizing)

### Technical
- Built with Go 1.21+
- Uses Cobra for CLI framework
- OpenAPI-driven code generation with oapi-codegen
- Secure template system preventing injection attacks
- Comprehensive error handling and validation
- Retry logic with exponential backoff
- Configuration via YAML, environment variables, or flags

## [0.9.0] - 2025-01-08

### Added
- Initial project structure
- Basic API integration
- Core command implementations
- Configuration management system

## Notes

### Version Numbering
- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality additions
- PATCH version for backwards-compatible bug fixes

### Future Releases
Future versions may include:
- Support for additional API specifications
- Plugin system for custom commands
- Web UI for visual workflow creation
- Integration with CI/CD pipelines
- Enhanced reporting and analytics
