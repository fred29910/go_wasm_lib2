## Context

The go_wasm_lib2 project generates type-safe Go and TypeScript SDKs from OpenAPI 3.x specifications, compiling the Go runtime to WebAssembly for browser use. Current implementation parses OpenAPI specs, builds an internal model, and renders templates for both languages.

**Current limitations:**
- TypeScript enums generated as `string` instead of union types (`'available' | 'pending' | 'sold'`)
- Go field names use `Id` instead of `ID`, `Url` instead of `URL`, etc.
- Only 200 response generates typed structs; other codes fall back to `interface{}`/`any`
- No validation of required fields in generated code
- Templates are embedded and not customizable
- CLI lacks dry-run, verbose modes, and structured error output

## Goals / Non-Goals

**Goals:**
1. Generate proper TypeScript union types for OpenAPI `enum` fields
2. Apply Go idiomatic naming conventions (ID, URL, HTTP, JSON, etc.)
3. Generate typed response structs for all documented HTTP status codes
4. Add optional request validation in generated Go code
5. Support custom templates via file paths
6. Improve CLI with dry-run, verbose, and JSON output modes
7. Add comprehensive unit tests for type mapping and generation

**Non-Goals:**
- Full OpenAPI 3.1 support (oneOf/anyOf/allOf) - too complex for this change
- External $ref resolution - architectural change
- Code generation for other languages (Python, Java, etc.)
- WASM runtime feature changes beyond supporting new generated patterns
- Breaking changes to existing CLI interface

## Decisions

### 1. Enum Handling Strategy

**Decision:** Parse `enum` values from OpenAPI Schema and generate:
- TypeScript: Union type `type Status = 'available' | 'pending' | 'sold'`
- Go: Custom type with validation `type Status string` + `Validate() error`

**Rationale:** 
- TypeScript union types provide compile-time safety and IDE autocomplete
- Go custom types with validation maintain type safety without external deps
- Alternatives considered: 
  - String constants in Go (less type-safe)
  - Generated constants + string type (no validation)
  - Using `go-enum` library (adds dependency)

### 2. Go Naming Conventions

**Decision:** Apply acronym handling in `ToGoName`:
- `id` → `ID`, `url` → `URL`, `http` → `HTTP`, `json` → `JSON`, `api` → `API`
- Apply to struct fields, type names, and method names

**Rationale:** Follows Go community conventions (golint, go vet). Current `Id`/`Url` triggers linter warnings.

### 3. Multi-Response Generation

**Decision:** For each operation, generate:
- Primary response (2xx, typically 200/201) → typed struct
- Error responses (4xx, 5xx) → typed error structs if schema provided
- Fallback: `interface{}`/`any` for undocumented responses

**Rationale:** Most APIs document error responses. Typed errors enable better error handling in generated code.

### 4. Request Validation

**Decision:** Add optional `Validate() error` method to request structs:
- Check required fields are non-zero/non-nil
- Validate enum values against allowed set
- Validate format constraints (email, uuid, date-time)

**Rationale:** Catches errors before network call. Optional to avoid breaking existing code.

### 5. Custom Template Support

**Decision:** Add `--go-template` and `--ts-template` flags accepting file paths. Fall back to embedded templates if not provided.

**Rationale:** Allows users to customize output without forking. Minimal API surface.

### 6. CLI Enhancements

**Decision:** Add flags:
- `--dry-run` - Parse and validate spec, show what would be generated without writing files
- `--verbose` - Show detailed progress (parsing, model building, template rendering)
- `--output json` - Machine-readable output for CI/CD integration

**Rationale:** Standard CLI patterns, improves CI/CD integration and debugging.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Enum generation breaks existing TypeScript consumers | Union types are compatible with string assignments; existing code continues to work |
| Go naming changes break existing Go consumers | Non-breaking for new fields; existing generated code unaffected. Add migration note in README |
| Template customization increases maintenance burden | Document template variables clearly; provide example templates |
| Validation adds runtime overhead | Make validation optional via build tag or config flag |
| Multi-response types increase generated code size | Only generate for documented responses; use shared error types where possible |

## Migration Plan

1. Implement enum support in type mapping (`types.go`)
2. Update Go naming in `ToGoName` function
3. Extend model building to capture all responses
4. Update templates to render new types
5. Add validation methods to generated structs
6. Add CLI flags and dry-run logic
7. Write comprehensive tests
8. Update documentation and examples

## Open Questions

1. Should validation be opt-in (flag) or opt-out (build tag)?
2. How to handle conflicting operation IDs across different paths with same method?
3. Should we generate separate error types per operation or shared error types per status code?
4. For custom templates: support Go template syntax only, or allow pluggable renderers?