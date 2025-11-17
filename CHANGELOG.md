# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of tq
- TOON format encoder and decoder
- jq-compatible query engine
- Format conversion (JSON, YAML, TOON)
- CLI with comprehensive flags
- Basic query operations:
  - Identity (`.`)
  - Field access (`.field`, `.nested.field`)
  - Array indexing (`.[0]`, `.[-1]`)
  - Array iteration (`.[]`)
  - Pipe operations (`|`)
  - Select/filter (`select()`)
  - Object construction (`{key: value}`)
  - Array construction (`[expr]`)
- Token usage statistics (`--stats`)
- Support for tabular arrays (uniform objects)
- Support for primitive arrays
- Support for nested objects
- Pretty-printing and compact output
- Multiple input/output formats
- Example files and documentation
- Comprehensive README
- MIT License
- GitHub Actions CI/CD
- Taskfile for common operations
- Unit tests for core functionality

### Documentation
- README with usage examples
- CONTRIBUTING guide
- EXAMPLES with real-world use cases
- Inline code documentation

### Infrastructure
- Go module setup
- GitHub workflows (test, release)
- Taskfile for build automation
- Example data files
- .gitignore configuration

## [0.1.0] - TBD

Initial alpha release.

### Features
- TOON format support
- Basic query engine
- Format conversion
- CLI interface

---

## Release Process

1. Update CHANGELOG.md with release notes
2. Update version in code
3. Create git tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
4. Push tag: `git push origin v0.1.0`
5. GitHub Actions will automatically build and create release
