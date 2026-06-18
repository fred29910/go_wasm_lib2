# Contributing to go_wasm_lib2

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/gowasm.git`
3. Create a feature branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `make test-race`
6. Commit and push: `git push origin feature/your-feature`
7. Open a Pull Request

## Development Setup

```bash
# Prerequisites: Go 1.24+, Node.js 18+ (for oxlint), TinyGo 0.41+ (optional)

# Install dependencies
make deps

# Run all tests with race detection
make test-race

# Run tests with coverage
make test-cover

# Full verification
make verify
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- All exported functions must have doc comments
- Use `log/slog` for structured logging (no `fmt.Print`)
- Keep functions focused and under 50 lines when possible
- Add tests for new functionality

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `security`

Examples:
```
feat(generator): add oneOf/anyOf/allOf schema support
fix(runtime): resolve nil pointer in ResolvePath
test(validator): add email validation edge case tests
docs(architecture): update package structure diagram
security(client): add path traversal protection
```

## Pull Request Process

1. Update documentation for any changed behavior
2. Add tests for new features or bug fixes
3. Ensure `make test-race` passes
4. Update CHANGELOG.md under `[Unchanged]`
5. Request review from maintainers

## Reporting Issues

- Use the [issue templates](.github/ISSUE_TEMPLATE/)
- Include steps to reproduce for bugs
- Include expected vs actual behavior
- Include Go version and OS information

## Code Review

All submissions require review. We use GitHub pull requests for this purpose.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
