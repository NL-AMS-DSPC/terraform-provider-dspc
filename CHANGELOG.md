# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of DSPC Terraform Provider
- Support for creating and deleting virtual machines
- Support for listing virtual machines via data source
- API key authentication support (Bearer token)
- Environment variable configuration support
- Unit tests with mocked API responses
- GitHub Actions workflow for automated releases
- Goreleaser configuration for multi-platform builds

### Security
- API key is marked as sensitive in provider configuration
- Authentication headers sent with all API requests (preparing for future API auth)

## [1.0.0] - TBD

### Added
- Initial release
- Basic VM resource management (create, read, delete)
- VM data source for listing all VMs
- Provider configuration with endpoint, timeout, and API key
- Support for Linux, Windows, and macOS (amd64/arm64)
- Terraform Registry publishing workflow

### Notes
- This version supports the minimal DSPC VM API (vmName field only)
- Future versions will add support for additional VM fields (cpu, memory, disk, etc.) as the API evolves
- Authentication is implemented but the current DSPC API doesn't validate API keys yet

## Versioning Strategy

This provider follows semantic versioning (SemVer):

- **MAJOR** version increments for incompatible API changes
- **MINOR** version increments for new functionality in a backwards compatible manner  
- **PATCH** version increments for backwards compatible bug fixes

### API Compatibility

- **v1.x.x**: Supports minimal VM API (name field only)
- **v2.x.x**: Will support extended VM API (cpu, memory, disk, etc.) when available
- **v3.x.x**: Will support additional resource types (containers, storage, etc.)

### Breaking Changes

Breaking changes will be documented in the CHANGELOG and will trigger a major version increment. Examples of breaking changes:

- Removing or renaming provider configuration options
- Changing resource attribute names or types
- Modifying API endpoint paths or request/response formats
- Changing authentication mechanisms
