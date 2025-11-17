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

## [0.2.0] - 2025-01-17

Major feature release with comprehensive built-in function library, advanced query features, and token comparison tool.

### Added - Query Engine Functions
- **String utilities**: `tostring()`, `tonumber()`, `ltrimstr()`, `rtrimstr()`
- **Object transformations**: `to_entries()`, `from_entries()`, `with_entries()`
- **Math functions**: `add()`, `min()`, `max()`, `floor()`, `ceil()`, `round()`
- **Array utilities**: `unique()`, `flatten()`, `flatten(depth)`, `range()`, `first()`, `first(n)`, `last()`, `last(n)`
- **Range generation**: `range(n)`, `range(from;to)`, `range(from;to;step)`
- **34 built-in functions** total (up from 16)

### Added - Query Syntax Features
- **Object construction**: `{key: value}` and shorthand `{name, age}`
- **Array construction**: `[expr]` for collecting results
- **Complex expressions**: Support for `[.users[] | {name, age}]`
- **Nested object/array construction** with proper precedence handling

### Added - CLI Features
- **`--slurp` mode**: Read multiple JSON/YAML/TOON values into a single array
- **`--null-input` mode**: Run queries without reading input (useful with `range()` and generators)
- **`--compare` flag**: Show format comparison (JSON/YAML/TOON sizes and token savings)
  - Displays all three format sizes side-by-side
  - Shows estimated token counts for each format
  - Calculates percentage savings or increase
  - Perfect for demonstrating TOON's token efficiency

### Improved
- **Parser precedence**: Fixed order of operations for construction operators
- **Query engine**: Better handling of nested structures and complex expressions
- **Error messages**: More descriptive errors for query parsing issues
- Comprehensive test coverage for all built-in functions
- Updated documentation with all function examples

### Fixed
- Array construction with pipes now works correctly (`[.users[] | select(.active)]`)
- Object construction respects nested structures in comma splitting
- Precedence issue where pipes inside brackets were parsed incorrectly

### Known Limitations
- Deep nesting with parentheses in object construction may cause stack overflow
- Boolean field access in `select()` requires comparison operators (use `.active == true`)

### Tests
- Added 7 string/object transformation function tests
- Added 6 math function tests (add, min, max, floor, ceil, round)
- Added 10 array utility tests (unique, flatten variants, range variants, first/last variants)
- Added object/array construction integration tests
- All 30+ tests pass

## [0.1.0] - 2025-01-17

Initial alpha release.

### Features
- TOON format support with encoder/decoder
- Basic query engine with 16 essential functions
- Format conversion (JSON ↔ YAML ↔ TOON)
- CLI interface with comprehensive flags
- Field access, array indexing, array iteration
- Pipe operations and select/filter
- Core functions: length, keys, values, type, sort, reverse
- Array operations: map, sort_by, group_by
- String functions: split, join, startswith, endswith, contains
- Object functions: has, in

---

## Release Process

1. Update CHANGELOG.md with release notes
2. Update version in code
3. Create git tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
4. Push tag: `git push origin v0.1.0`
5. GitHub Actions will automatically build and create release
