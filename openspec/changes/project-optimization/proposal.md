## Why

The current go_wasm_lib2 project generates Go and TypeScript SDKs from OpenAPI specs but has several limitations that reduce developer experience and production readiness. The generated TypeScript lacks proper enum support, Go code doesn't follow naming conventions, and key OpenAPI features (oneOf/anyOf, validation, multiple response codes) are unsupported. These gaps require manual post-generation fixes and limit adoption.

## What Changes

- **Add TypeScript enum generation** - Generate proper union types for OpenAPI enum fields (e.g., `status: 'available' | 'pending' | 'sold'`)
- **Fix Go naming conventions** - Use `ID` instead of `Id`, proper PascalCase for all generated types
- **Add response type generation** - Generate typed response structs/interfaces for all response codes, not just 200
- **Add request/response validation** - Generate validation logic for required fields and constraints
- **Improve template flexibility** - Support custom templates and partial generation
- **Add comprehensive test coverage** - Unit tests for edge cases in type mapping, enum handling, and code generation
- **Enhance CLI usability** - Add dry-run mode, better error messages, and progress reporting
- **Update oxlint configuration** - Align with generated code patterns to avoid false positives

## Capabilities

### New Capabilities
- `typescript-enum-support`: Generate proper TypeScript union types from OpenAPI enum fields
- `go-naming-conventions`: Apply Go idiomatic naming (ID, URL, HTTP, etc.) to generated code
- `multi-response-support`: Generate typed responses for all HTTP status codes in OpenAPI spec
- `request-validation`: Generate runtime validation for required fields and constraints
- `custom-template-support`: Allow users to provide custom templates for code generation
- `cli-enhancements`: Dry-run mode, verbose output, better error handling

### Modified Capabilities
- `openapi-codegen`: Enhanced to support enums, validation, multiple responses, and improved naming
- `wasm-runtime`: Minor updates to support new generated code patterns

## Impact

**Affected code:**
- `pkg/generator/generator.go` - Core generation logic, type mapping, model building
- `pkg/generator/openapi.go` - OpenAPI parsing, enum extraction
- `pkg/generator/types.go` - Type conversion functions (Go/TypeScript)
- `pkg/generator/templates/sdk.go.tmpl` - Go template
- `pkg/generator/templates/sdk.ts.tmpl` - TypeScript template
- `cmd/generator/main.go` - CLI flags and commands
- `oxlintrc.json` - Linting rules for generated TypeScript

**Dependencies:**
- No new external dependencies required
- Existing: `gopkg.in/yaml.v3`, `github.com/norunners/vert`, `github.com/urfave/cli/v2`