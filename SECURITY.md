# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 1.0.x   | ✅ Current         |

## Reporting a Vulnerability

Please report security vulnerabilities by opening a [GitHub Issue](https://github.com/fred29910/gowasm/issues) with the `security` label.

We will acknowledge receipt within 48 hours and provide a fix timeline.

## Security Features

The project implements the following security measures:

### Input Validation
- **Path Traversal Protection**: `safePathParam` rejects path parameters containing `..`, `//`, or absolute paths
- **Response Body Limit**: 10MB `LimitReader` prevents OOM attacks
- **URL Safe Construction**: `url.JoinPath` prevents double-slash and path truncation

### JavaScript Interop Security
- **Prototype Pollution Protection**: 13 dangerous JavaScript keys are filtered during JS↔Go conversion
- **Promise Panic Recovery**: All Promise executors have `recover()` protection

### Concurrency Safety
- `HTTPClient.mu` (RWMutex) protects `config.Headers`
- `operationsMu` (RWMutex) protects the operation registry
- `WASMExports.mu` (RWMutex) protects initialization state

### Error Handling
- Structured `WASMError` with automatic caller location capture
- `Unwrap()` support for `errors.Is`/`errors.As` compatibility
- No sensitive data in error messages

## Dependency Security

- Dependencies are monitored via Dependabot
- All dependencies are pinned in `go.sum`
- Indirect dependencies are minimized
