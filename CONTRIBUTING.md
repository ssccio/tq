# Contributing to tq

Thank you for your interest in contributing to `tq`! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Task (optional, for using Taskfile)
- Git

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/tq.git
   cd tq
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the project:
   ```bash
   task build
   # or
   go build -o tq cmd/tq/main.go
   ```

5. Run tests:
   ```bash
   task test
   # or
   go test ./...
   ```

## Project Structure

```
tq/
├── cmd/tq/              # CLI entry point
├── pkg/
│   ├── toon/            # TOON format encoding/decoding
│   ├── query/           # Query engine (jq-compatible)
│   ├── converter/       # Format converters (JSON/YAML/TOON)
│   └── cli/             # CLI command handling
├── internal/            # Internal packages
├── examples/            # Example files
├── tests/               # Integration tests
└── docs/                # Additional documentation
```

## Development Workflow

### Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the code style guidelines

3. Add tests for your changes

4. Run tests and linting:
   ```bash
   task test
   task lint
   ```

5. Commit your changes:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

Examples:
```
feat: add support for streaming large files
fix: correct array index handling in query engine
docs: update README with new examples
test: add unit tests for TOON decoder
```

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized

### Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage (>80%)
- Include both positive and negative test cases

Run tests:
```bash
# Unit tests
go test ./...

# With coverage
task test-coverage

# Specific package
go test ./pkg/toon -v
```

## Pull Request Process

1. Update the README.md or documentation if needed
2. Add tests for your changes
3. Ensure all tests pass
4. Update CHANGELOG.md with your changes
5. Push to your fork and submit a pull request

### Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Include examples if adding new features
- Ensure CI checks pass
- Be responsive to feedback

## Feature Requests and Bug Reports

### Reporting Bugs

When reporting bugs, please include:

- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Sample data if applicable

### Requesting Features

For feature requests:

- Describe the use case
- Explain why it would be useful
- Provide examples if possible
- Consider implementation complexity

## Code Review Process

- Maintainers will review PRs as soon as possible
- Address review feedback promptly
- Be open to suggestions and constructive criticism
- Small, focused PRs are easier to review

## Areas for Contribution

Looking for ways to contribute? Here are some areas:

### High Priority
- [ ] Full jq syntax compatibility
- [ ] Performance optimizations
- [ ] Streaming support for large files
- [ ] Better error messages

### Medium Priority
- [ ] Interactive mode (like ijq)
- [ ] Shell completions (bash, zsh, fish)
- [ ] More output formats
- [ ] Plugin system

### Good First Issues
- [ ] Add more examples
- [ ] Improve documentation
- [ ] Add benchmarks
- [ ] Write integration tests

## Questions?

- Open an issue for discussion
- Check existing issues and PRs
- Read the documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to tq! 🎉
