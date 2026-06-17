# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Comprehensive code review report v2.0 (`docs/reviews/comprehensive-review-v2-2026-06-17.md`)
- MIT LICENSE file
- Unit tests for `pkg/runtime/validate` (email, UUID, datetime, enum, IsValid)
- Unit tests for `pkg/runtime/errors` (WASMError, WrapWASMError, FromError, predefined errors)
- Template pre-compilation with `sync.Once` for all built-in templates
- Makefile `test-race` target for race condition detection

### Changed
- **Validator**: Fixed `IsValidEmail` to support multi-dot domains (e.g. `user@mail.example.com`)
- **Validator**: Fixed `IsValid` to correctly detect nil slices/maps/channels via `reflect`
- **Validator**: `go.mod` lowered minimum Go version from 1.25.1 to 1.24
- **Generator**: All built-in templates now pre-compiled via `sync.Once` for better performance
- **CLI**: `--validation` flag no longer has `-v` alias (conflicted with `--version`)
- **CLI**: `--wasm` flag default changed from `true` to `false` (safer default)
- **CI**: Added `test-race` verification step

### Security
- Email validator bug could incorrectly reject valid addresses with subdomain dots
- Nil slice/map values could pass `IsValid` check

---

## [1.0.0] - 2026-06-17

### Initial Release

#### Core Features
- **OpenAPI 3.x SDK Generator**: Parse OpenAPI YAML/JSON specs and generate type-safe SDKs
- **Multi-language Output**: Generate Go client code, TypeScript SDK, WASM entry point, and interactive demo HTML
- **WASM Compilation**: Build WASM binaries with standard Go (2-5MB) or TinyGo (200-500KB)
- **HTTP Client Runtime**: Full-featured HTTP client compiled to WebAssembly with Promise-based async API
- **Authentication**: Bearer Token and custom scheme authentication support
- **Integrated Linting**: oxlint support for generated TypeScript code
- **Dry Run Mode**: Preview generated files without writing to disk

#### Code Generator
- Generate Go client code with:
  - Schema struct definitions with JSON tags
  - Pointer-type required fields with nil validation
  - Request/Response type definitions
  - APIClient with multi-instance support (no global singleton)
  - Validation methods (required check, enum validation, format validation)
  - ToRequest conversion functions with pointer dereferencing
  - Helper functions for type conversion
- Generate TypeScript SDK with:
  - Type-safe interfaces for all schemas and request/response types
  - `WASMSDK` class with `load()`/`init()`/`callAPI()`/`auth`/`getConfig()`
  - Typed API functions
  - Query parameters support multi-value types
- Generate `go.mod` with configurable module name
- Generate `cmd/wasm/main.go` with `//go:build js && wasm` build tag
- Generate interactive demo HTML page (Tailwind CSS)

#### CLI
- `generate` command with flags:
  - `-s, --spec`: OpenAPI spec path (required)
  - `-o, --out`: Output directory (default: `./generated`)
  - `-m, --module`: Go module name
  - `-p, --package`: Go package name
  - `-V, --verbose`: Detailed progress output
  - `--validation`: Generate validation methods (default: true)
  - `--wasm`: Build WASM after generation (default: false)
  - `--wasm-out`: WASM output path
  - `--compiler`: WASM compiler (auto/tinygo/go, default: auto)
  - `--go-template`: Custom Go template path
  - `--ts-template`: Custom TypeScript template path
  - `--oxlintrc`: Custom oxlint config path
  - `--oxlint-disable`: Disable oxlint
  - `--dry-run`: Preview mode
  - `--output`: Output format (text/json, default: text)
- `init` command: Create sample project structure
- JSON output mode for CI/CD integration
- Unified error output via `ExitErrHandler`

#### WASM Runtime
- HTTP client with:
  - Safe URL construction (`url.JoinPath`)
  - Path parameter replacement with regex (deterministic, no map-order injection)
  - Path traversal protection (`safePathParam`)
  - Response body size limit (10MB `LimitReader`)
  - Concurrent-safe headers (`sync.RWMutex`)
  - Automatic caller location capture (`runtime.Caller`)
- JavaScript interop:
  - Promise-based async API with recover protection
  - Go↔JavaScript type conversion (`vert` library)
  - Prototype pollution protection (13 dangerous keys filtered)
  - Structured WASMError with code/message/details/suggestion/filePath/lineNumber
  - `errors.Is`/`errors.As` support via `Unwrap()`
- Build orchestration:
  - Compiler auto-detection (tinygo → go fallback)
  - `go mod tidy` before build
  - Context-based timeout (5min build, 2min tidy)

#### Project Infrastructure
- Dual build system: GNU Make + Taskfile
- GitHub Actions CI/CD workflow (lint, test, build, generate, release)
- Dependabot configuration
- Issue and PR templates
- Branch protection documentation

#### Documentation
- `docs/architecture.md`: System architecture with Mermaid diagrams
- `docs/getting-started.md`: Setup guide and build system comparison
- `docs/cli-reference.md`: CLI command and flag reference
- `docs/runtime-api.md`: WASM exported functions and error codes
- `docs/generator-api.md`: Template system and custom template variables
- `docs/known-issues.md`: Known limitations and roadmap
- Two comprehensive code review reports

#### Security
- Path traversal prevention (safePathParam with `..`, `//`, `/` rejection)
- Prototype pollution protection (13 dangerous JS keys filtered)
- Response body OOM protection (10MB LimitReader)
- Concurrent access safety (3 RWMutex-protected shared states)
- Deterministic path parameter replacement (regex, not map iteration)
- Deterministic primary response selection (sorted codes)
