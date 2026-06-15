# go_wasm_lib2 - Generic WASM HTTP SDK Generator

A Go-based toolkit for building WebAssembly (WASM) HTTP clients and generating type-safe SDKs from OpenAPI 3.x specifications.

## Overview

This project provides two main components:

1. **WASM Runtime Library** (`pkg/runtime`) - A JavaScript-compatible runtime library compiled to WebAssembly, providing HTTP client capabilities with Promise-based API for browser environments. Includes:
   - HTTP client with timeout, headers, and authentication support
   - JavaScript ↔ Go type conversion (via `vert` library)
   - Promise-based async API for non-blocking operations
   - Thread-safe state management with `sync.RWMutex`
   - Structured error handling with error codes

2. **SDK Generator** (`pkg/generator`) - A code generator that creates type-safe HTTP client SDKs from OpenAPI 3.x specifications. Generates:
   - **Go client code** - Structs for request/response types, operation handlers, and runtime registration
   - **TypeScript SDK** - Interface definitions, typed API functions, and WASM wrapper class
   - **Demo HTML page** - Interactive testing interface with Tailwind CSS

## Features

### Runtime Features
- Build WebAssembly HTTP clients with standard Go compiler
- Build smaller WebAssembly binaries with TinyGo compiler (200-500KB vs 2-5MB)
- Promise-based async API compatible with JavaScript
- Thread-safe runtime with `sync.RWMutex` protection
- Bearer token authentication support
- Configurable timeout, headers, and credentials
- Structured error handling with error codes (`INVALID_CONFIG`, `TIMEOUT`, etc.)

### Generator Features
- Generate Go client code for WASM runtime with request/response structs
- Generate TypeScript SDK with typed interfaces and API functions
- Generate demo HTML page for interactive testing
- **Auto-compile WASM** after code generation with automatic tinyGo detection
- Support for OpenAPI 3.x specifications (schemas, operations, parameters, requestBody, responses)
- Automatic path and query parameter handling
- Type-safe request/response conversion (Go ↔ TypeScript)
- Customizable output (module name, package name, output directory)
- Configurable WASM compiler (auto/tinygo/go) via `--compiler` flag

## Quick Start

### Prerequisites

- Go 1.25.1 or later
- TinyGo (optional, for smaller WASM builds)

### Build the Runtime

```bash
# Build with standard Go compiler (larger, more features)
make build

# Build with TinyGo compiler (smaller, faster startup)
make build-tinygo

# Build both
make build-all
```

### Generate SDKs from OpenAPI Specs

```bash
# Generate SDK from your OpenAPI spec (auto-compiles WASM)
make generate SPEC=path/to/openapi.yaml OUT=./output

# Example with petstore
make dev-generate

# Using the CLI directly
gowasm-generator generate -s path/to/openapi.yaml -o ./output

# Force tinygo for WASM compilation
gowasm-generator generate -s path/to/openapi.yaml -o ./output --compiler tinygo
```

### Run Tests

```bash
# Test compilation
make test-compile

# Run Go tests
make test

# Full verification (deps, build, compile, generate)
make verify
```

## Usage

### Runtime Usage (WASM in Browser)

After building the WASM binary (`build/main.wasm`), use it in your web application:

```html
<!DOCTYPE html>
<html>
<head>
  <title>WASM API Client</title>
</head>
<body>
  <script type="module">
    // Load the WASM module
    const response = await fetch('./main.wasm');
    const bytes = await response.arrayBuffer();
    const { instance } = await WebAssembly.instantiate(bytes);
    
    // The WASM module exports functions to the global scope
    // After instantiation, these functions are available:
    // - window.wasmInitClient(config)
    // - window.wasmCallAPI(operationId, request)
    // - window.wasmSetAuthToken(token, scheme?)
    // - window.wasmClearAuthToken()
    // - window.wasmGetConfig()
    
    // Initialize the HTTP client
    await window.wasmInitClient({
      baseUrl: 'https://api.example.com',
      timeout: 30,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin'  // 'include' | 'omit' | 'same-origin'
    });
    
    // Set authentication token (optional)
    await window.wasmSetAuthToken('your-jwt-token', 'Bearer');
    
    // Make API calls
    const apiResponse = await window.wasmCallAPI('getUsers', {
      method: 'GET',
      path: '/users',
      headers: { 'Accept': 'application/json' },
      query: { page: '1', limit: '10' }
    });
    
    console.log('Status:', apiResponse.status);
    console.log('Body:', apiResponse.body);
    
    // Clear auth token when needed
    window.wasmClearAuthToken();
  </script>
</body>
</html>
```

**Exported WASM Functions:**

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `wasmInitClient(config)` | `config: WASMConfig` | `Promise<{success, message}>` | Initialize HTTP client |
| `wasmCallAPI(operationId, request)` | `operationId: string, request: HTTPRequest` | `Promise<HTTPResponse>` | Make API call |
| `wasmSetAuthToken(token, scheme?)` | `token: string, scheme?: string` | `{success, error?}` | Set auth header |
| `wasmClearAuthToken()` | none | `{success, error?}` | Remove auth header |
| `wasmGetConfig()` | none | `{success, config?, error?}` | Get current config |

**WASMConfig Interface:**

```typescript
interface WASMConfig {
  baseUrl: string;           // Base URL for API requests
  timeout?: number;          // Request timeout in seconds (default: 30)
  headers?: Record<string, string>;  // Default headers
  credentials?: 'include' | 'omit' | 'same-origin';  // Fetch credentials mode
}
```

**HTTPRequest Interface:**

```typescript
interface HTTPRequest {
  method: string;            // HTTP method (GET, POST, PUT, DELETE, PATCH)
  path: string;              // Request path (e.g., '/users/{id}')
  pathParams?: Record<string, string>;  // Path parameters
  headers?: Record<string, string>;     // Request headers
  query?: Record<string, string>;       // Query parameters
  body?: any;                // Request body (for POST/PUT/PATCH)
}
```

### Generated SDK Usage

After generating an SDK from an OpenAPI spec, use the generated TypeScript client:

```typescript
import { WASMSDK, Pet, CreatePetRequest, GetPetByIdRequest } from './generated/sdk';

// Create SDK instance
const sdk = new WASMSDK('./main.wasm');

// Load and initialize WASM
await sdk.load();
await sdk.init({ baseUrl: 'https://petstore3.swagger.io/api/v3' });

// Use typed API functions
const response = await sdk.createPet({
  body: { name: 'Fluffy', status: 'available' }
});

const pet: Pet = response.body as Pet;
console.log('Created pet:', pet.name);

// Or use the generic callAPI method
const getResponse = await sdk.callAPI('getPetById', {
  method: 'GET',
  path: '/pet/{petId}',
  pathParams: { petId: '123' }
});
```

**Generated TypeScript Types:**

```typescript
// Schema interfaces
interface Pet {
  id?: number;
  name: string;
  status?: 'available' | 'pending' | 'sold';
}

// Request interfaces
interface CreatePetRequest {
  body?: Pet;
}

interface GetPetByIdRequest {
  petId: number;
  pathParams?: Record<string, string>;
}

// API functions (auto-generated)
function createPet(params: CreatePetRequest): Promise<HTTPResponse>;
function getPetById(params: GetPetByIdRequest): Promise<HTTPResponse>;
function findPetsByStatus(params: FindPetsByStatusRequest): Promise<HTTPResponse>;
```

## Project Structure

```
.
├── cmd/
│   ├── generator/                    # SDK generator CLI entry point
│   │   └── main.go                  # CLI flags: -spec, -out, -module, -package
│   └── runtime/                      # WASM runtime entry point
│       └── main.go                  # Build constraint: js && wasm
├── pkg/
│   ├── generator/                    # Core generation logic
│   │   ├── generator.go            # Main generator: model building, orchestration
│   │   ├── openapi.go              # OpenAPI 3.x YAML parser with validation
│   │   ├── types.go                # Type definitions and naming conversions
│   │   ├── go_templates.go         # Go code generation templates (go:embed)
│   │   ├── ts_templates.go         # TypeScript + demo HTML templates (go:embed)
│   │   ├── templates/              # Embedded template files
│   │   │   ├── sdk.go.tmpl         # Go client template
│   │   │   ├── go.mod.tmpl         # Go module template
│   │   │   ├── sdk.ts.tmpl         # TypeScript SDK template
│   │   │   └── index.html.tmpl     # Demo HTML template
│   │   ├── generator_test.go       # Generator unit tests
│   │   ├── openapi_test.go         # OpenAPI parser tests
│   │   └── types_test.go           # Type conversion tests
│   └── runtime/                      # WASM runtime core
│       ├── build.go                # WASM build utility (auto-detect tinygo)
│       ├── build_test.go           # WASM build tests
│       ├── client.go               # HTTP client, request/response, operation registry
│       ├── exports.go              # JavaScript-callable WASM exports
│       ├── promise.go              # Promise helper for async JS interop
│       ├── converter.go            # Go ↔ JavaScript type conversion via vert
│       └── error.go                # Structured error types with error codes
├── examples/
│   ├── petstore/
│   │   └── openapi.yaml             # Sample OpenAPI spec (Petstore API)
│   └── templates/
│       ├── custom.go.tmpl           # Custom Go template example
│       └── custom.ts.tmpl           # Custom TypeScript template example
├── build/                            # Generated WASM binaries (gitignored)
├── Makefile                          # GNU Make build system (81 lines)
├── Taskfile.yml                      # Task runner build system (185 lines)
├── go.mod                            # Go module definition
├── reviews.md                        # Code review report
└── 修复摘要.md                        # Bug fix summary (Chinese)
```

## Configuration

### Generator CLI Options

```bash
gowasm-generator generate \
  -s path/to/openapi.yaml \
  -o ./output \
  -m mydomain \
  -p client
```

| Flag | Aliases | Description | Default | Required |
|------|---------|-------------|---------|----------|
| `--spec` | `-s` | Path to OpenAPI YAML specification | - | Yes |
| `--out` | `-o` | Output directory for generated code | `./generated` | No |
| `--module` | `-m` | Go module name for generated code | `github.com/fred29910/gowasm` | No |
| `--package` | `-p` | Package name for generated Go code | `generated` | No |
| `--wasm` | | Build WASM after generation | `true` | No |
| `--wasm-out` | | Output path for WASM binary | `<out>/main.wasm` | No |
| `--compiler` | | WASM compiler: `auto`, `tinygo`, `go` | `auto` | No |
| `--oxlintrc` | | Path to custom oxlintrc.json | embedded | No |
| `--oxlint-disable` | | Disable oxlint after generation | `false` | No |
| `--validation` | `-v` | Generate validation methods for request structs | `true` | No |
| `--go-template` | | Path to custom Go template file | embedded | No |
| `--ts-template` | | Path to custom TypeScript template file | embedded | No |

**WASM Compilation:**

The generator automatically compiles WASM after generating code. It detects `tinygo` in PATH and prefers it for smaller binaries:

```bash
# Default: auto-detect compiler (tinygo > go)
gowasm-generator generate -s openapi.yaml -o ./output

# Force tinygo (fails if not installed)
gowasm-generator generate -s openapi.yaml -o ./output --compiler tinygo

# Force standard go
gowasm-generator generate -s openapi.yaml -o ./output --compiler go

# Custom WASM output path
gowasm-generator generate -s openapi.yaml -o ./output --wasm-out ./build/sdk.wasm

# Skip WASM compilation
gowasm-generator generate -s openapi.yaml -o ./output --wasm=false

# Skip validation method generation
gowasm-generator generate -s openapi.yaml -o ./output --validation=false
```

**Compiler Selection (`--compiler`):**

| Value | Behavior |
|-------|----------|
| `auto` | Use `tinygo` if available in PATH, otherwise fall back to `go` |
| `tinygo` | Require `tinygo` (error if not found) |
| `go` | Use standard `go` compiler with `GOOS=js GOARCH=wasm` |

**Example with all options:**

```bash
gowasm-generator generate \
  -s ./api/openapi.yaml \
  -o ./sdk/generated \
  -m github.com/myorg/myproject \
  -p apiclient \
  --compiler tinygo \
  --wasm-out ./build/sdk.wasm \
  --go-template ./templates/custom.go.tmpl \
  --ts-template ./templates/custom.ts.tmpl
```

**Generated files:**

| File | Description |
|------|-------------|
| `generated.go` | Go client code with request/response structs and operation handlers |
| `go.mod` | Go module definition for the generated SDK |
| `sdk.ts` | TypeScript SDK with interfaces, typed API functions, and WASM wrapper |
| `index.html` | Interactive demo page for testing the API |
| `main.wasm` | Compiled WASM binary (if `--wasm=true`) |

### Runtime Configuration

The WASM runtime accepts a configuration object when initializing the HTTP client:

```typescript
interface WASMConfig {
  baseUrl: string;           // Base URL for all API requests
  timeout?: number;          // Request timeout in seconds (default: 30)
  headers?: Record<string, string>;  // Default headers sent with every request
  credentials?: 'include' | 'omit' | 'same-origin';  // Fetch credentials mode
}
```

**Configuration examples:**

```javascript
// Basic configuration
await wasmInitClient({
  baseUrl: 'https://api.example.com'
});

// Full configuration
await wasmInitClient({
  baseUrl: 'https://api.example.com',
  timeout: 60,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key'
  },
  credentials: 'include'  // Send cookies cross-origin
});
```

### Custom Templates

The generator supports custom Go `text/template` files for both Go and TypeScript code generation. This allows you to customize the output format, add boilerplate, or generate code in a different style.

#### CLI Flags

```bash
# Use a custom Go template
gowasm-generator generate -s openapi.yaml -o ./output --go-template ./my-go.tmpl

# Use a custom TypeScript template
gowasm-generator generate -s openapi.yaml -o ./output --ts-template ./my-ts.tmpl

# Combine both
gowasm-generator generate -s openapi.yaml -o ./output \
  --go-template ./my-go.tmpl \
  --ts-template ./my-ts.tmpl
```

#### Template Variables

Templates use Go's `text/template` syntax. The following data structures are available:

**Root Data Structure (`GenerationModel`):**

```go
type GenerationModel struct {
    Doc           *OpenAPI              // Parsed OpenAPI document
    Config        *Config               // Generator configuration
    InfoTitle     string                // API title from OpenAPI info
    InfoVersion   string                // API version from OpenAPI info
    BaseURL       string                // Default server URL
    Schemas       []GeneratedSchema     // All schema definitions
    Operations    []GeneratedOperation  // All operation definitions
    OperationIDs  []string              // Sorted list of operation IDs
    Validation    bool                  // Whether validation methods are enabled
}
```

**Config (`{{.Config}}`):**

```go
type Config struct {
    ModuleName    string  // Go module name (e.g., "github.com/myorg/myproject")
    OutputModule  string  // Output module path
    Package       string  // Go package name (e.g., "generated")
    RuntimePath   string  // Local path to runtime module
    RuntimeImport string  // Import path of the runtime package
    Validation    bool    // Whether validation is enabled
}
```

**GeneratedSchema (`{{.Schemas}}`):**

```go
type GeneratedSchema struct {
    Name       string              // Original schema name from OpenAPI
    GoName     string              // Go identifier name (PascalCase)
    TSName     string              // TypeScript identifier name
    Properties []GeneratedProperty // Schema properties
}
```

**GeneratedProperty (`{{.Schemas.Properties}}`):**

```go
type GeneratedProperty struct {
    Name       string        // Original property name
    GoName     string        // Go field name
    TSName     string        // TypeScript field name
    GoType     string        // Go type (e.g., "string", "int64", "[]Pet")
    TSType     string        // TypeScript type (e.g., "string", "number", "Array<Pet>")
    Required   bool          // Whether the property is required
    EnumValues []interface{} // Allowed values (if enum)
    Format     string        // Format hint (e.g., "email", "uuid", "date-time")
}
```

**GeneratedOperation (`{{.Operations}}`):**

```go
type GeneratedOperation struct {
    ID                  string                // Operation ID from OpenAPI
    TSName              string                // TypeScript function name
    Method              string                // HTTP method (GET, POST, etc.)
    Path                string                // Request path (e.g., "/pet/{petId}")
    Summary             string                // Operation summary
    RequestStructName   string                // Go request struct name
    RequestInterface    string                // Go request interface name
    ResponseStructName  string                // Go response struct name
    ResponseInterface   string                // Go response interface name
    PathParams          []GeneratedParameter  // Path parameters
    QueryParams         []GeneratedParameter  // Query parameters
    HasBody             bool                  // Whether operation has request body
    BodyType            string                // Go type for request body
    ResponseType        string                // Go type for primary response
    BodyParamName       string                // Body parameter name (always "body")
    Responses           []GeneratedResponse   // All possible responses
}
```

**GeneratedParameter (`{{.Operations.PathParams}}` / `{{.Operations.QueryParams}}`):**

```go
type GeneratedParameter struct {
    Name       string        // Original parameter name
    GoName     string        // Go field name
    TSName     string        // TypeScript field name
    GoType     string        // Go type
    TSType     string        // TypeScript type
    In         string        // "path" or "query"
    Required   bool          // Whether parameter is required
    EnumValues []interface{} // Allowed values (if enum)
    Format     string        // Format hint
}
```

**GeneratedResponse (`{{.Operations.Responses}}`):**

```go
type GeneratedResponse struct {
    Code        string  // HTTP status code (e.g., "200", "404")
    GoType      string  // Go type for response body
    TSType      string  // TypeScript type for response body
    Description string  // Response description
    Primary     bool    // Whether this is the primary (2xx) response
    StructName  string  // Go response struct name
}
```

#### Template Functions

The following custom template functions are available:

| Function | Signature | Description |
|----------|-----------|-------------|
| `hasPrefix` | `hasPrefix(s, prefix string) bool` | Check if string has prefix |

#### Example: Custom Go Template

```go
// my-go.tmpl
package {{.Config.Package}}

import (
    "context"
    {{.Config.RuntimeImport}}
)

// API title: {{.InfoTitle}}
// API version: {{.InfoVersion}}

{{range .Schemas}}
type {{.GoName}} struct {
{{range .Properties}}    {{.GoName}} {{.GoType}} `json:"{{.Name}},omitempty"`
{{end}}}
{{end}}

{{range .Operations}}
func {{.RequestInterface}}ToRequest(params {{.RequestInterface}}) runtime.Request {
    return runtime.Request{
        Method: "{{.Method}}",
        Path:   "{{.Path}}",
{{if .HasBody}}        Body:   params.Body,
{{end}}    }
}
{{end}}
```

#### Example: Custom TypeScript Template

```typescript
// my-ts.tmpl
// Generated from {{.InfoTitle}} v{{.InfoVersion}}

{{range .Schemas}}
export interface {{.GoName}} {
{{range .Properties}}    {{.TSName}}{{if not .Required}}?{{end}}: {{.TSType}};
{{end}}}
{{end}}

{{range .Operations}}
export async function {{.TSName}}(params: {{.RequestInterface}}): Promise<HTTPResponse> {
    return (window as any).wasmCallAPI('{{.ID}}', {
        method: '{{.Method}}',
        path: '{{.Path}}',
{{if .HasBody}}        body: params.body,
{{end}}    });
}
{{end}}
```

#### Example Templates

The `examples/templates/` directory contains ready-to-use custom template examples:

| Template | Description |
|----------|-------------|
| `examples/templates/custom.go.tmpl` | Custom Go template with section comments, detailed field documentation, and helper functions |
| `examples/templates/custom.ts.tmpl` | Custom TypeScript template with section comments, detailed JSDoc-style comments, and typed API functions |

To use these examples:

```bash
# Use the custom Go template
gowasm-generator generate -s openapi.yaml -o ./output --go-template ./examples/templates/custom.go.tmpl

# Use the custom TypeScript template
gowasm-generator generate -s openapi.yaml -o ./output --ts-template ./examples/templates/custom.ts.tmpl

# Use both custom templates
gowasm-generator generate -s openapi.yaml -o ./output \
  --go-template ./examples/templates/custom.go.tmpl \
  --ts-template ./examples/templates/custom.ts.tmpl
```

These examples demonstrate:
- **Section organization** with clear comment separators
- **Field documentation** including required/optional status and format hints
- **Available template variables** listed at the top of each file
- **Best practices** for template structure and formatting

### OpenAPI Spec Requirements

The generator expects OpenAPI 3.x specifications with the following structure:

```yaml
openapi: "3.0.0"
info:
  title: API Title
  version: "1.0.0"
servers:
  - url: "https://api.example.com/v1"
paths:
  /resource:
    get:
      operationId: getResource
      summary: Get resource
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Resource"
      responses:
        "200":
          description: Success
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Resource"
components:
  schemas:
    Resource:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
```

**Supported OpenAPI features:**

| Feature | Support | Notes |
|---------|---------|-------|
| Paths | ✅ Full | GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD |
| Parameters | ✅ Full | Path and query parameters |
| Request Body | ✅ Full | JSON request bodies |
| Responses | ✅ Full | JSON response parsing |
| Schemas | ✅ Full | Objects, arrays, primitives |
| $ref | ✅ Full | Internal references (`#/components/schemas/...`) |
| oneOf/anyOf/allOf | ⚠️ Partial | Not yet supported |
| External $ref | ❌ Not supported | Only internal references |

## Build Systems

This project supports two build systems. Choose the one you prefer:

### Make (GNU Make)

```bash
make help              # List all available targets
make build             # Build WASM binary with standard Go
make build-tinygo      # Build with TinyGo (smaller binary)
make build-all         # Build both versions
make generate SPEC=... OUT=...  # Generate SDK from OpenAPI spec
make dev-generate      # Generate petstore example SDK
make test-compile      # Test WASM compilation
make test              # Run Go unit tests
make verify            # Full verification (deps, build, compile, generate)
make deps              # Download and tidy dependencies
make clean             # Remove build artifacts
make clean:all         # Deep clean including generated files
make install-tinygo-macos   # Install TinyGo on macOS
make install-tinygo-linux   # Install TinyGo on Linux
```

### Task (Taskfile)

```bash
task                   # List all available tasks
task build             # Build WASM binary with standard Go
task build:tinygo      # Build with TinyGo (smaller binary)
task build:all         # Build both versions
task generate SPEC=... OUT=...  # Generate SDK from OpenAPI spec
task dev:generate      # Generate petstore example SDK
task test:compile      # Test WASM compilation
task test              # Run Go unit tests
task verify            # Full verification (deps, build, compile, generate)
task deps              # Download and tidy dependencies
task clean             # Remove build artifacts
task clean:all         # Deep clean including generated files
task install:tinygo:macos   # Install TinyGo on macOS
task install:tinygo:linux   # Install TinyGo on Linux
```

### Build System Comparison

| Feature | Make | Task |
|---------|------|------|
| Syntax | Tab-based, shell commands | YAML-based, cross-platform |
| Dependency | Requires `make` | Requires `task` |
| Source tracking | Limited | Built-in (`sources`, `generates`) |
| Cross-platform | Unix-focused | Full Windows support |
| Verbose output | Default | Configurable |

### Build Output

| Target | Command | Output | Size (approx) |
|--------|---------|--------|---------------|
| Standard Go | `make build` | `build/main.wasm` | 2-5 MB |
| TinyGo | `make build-tinygo` | `build/tinymain.wasm` | 200-500 KB |
| Auto (generate) | `gowasm-generator generate` | `<out>/main.wasm` | depends on compiler |

**Note:** TinyGo produces significantly smaller binaries but may have some compatibility limitations with certain Go packages.

**Auto-detection:** When using `gowasm-generator generate`, the `--compiler auto` mode checks for `tinygo` in PATH and uses it if available, otherwise falls back to standard `go`.

## Development

### Install TinyGo

```bash
# macOS (via Homebrew)
make install-tinygo-macos

# Linux (via .deb package)
make install-tinygo-linux

# Verify installation
tinygo version
```

### Development Workflow

```bash
# Full development setup (deps, build, generate example)
make dev

# Clean everything including generated files
make clean:all

# Run full verification suite
make verify
```

### Code Structure

- **Generator** (`pkg/generator/`): Parse OpenAPI specs and generate client code
- **Runtime** (`pkg/runtime/`): WASM runtime with HTTP client and JS interop
- **Templates** (`go_templates.go`, `ts_templates.go`): Code generation templates
- **Examples** (`examples/`): Sample OpenAPI specs and generated output

## API Reference

### Go Runtime API

**Client Configuration:**

```go
// DefaultClientConfig returns default configuration
func DefaultClientConfig() *ClientConfig

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *ClientConfig) *HTTPClient

// SetDefaultClient replaces the default client
func SetDefaultClient(config *ClientConfig)
```

**Operation Registry:**

```go
// RegisterOperation registers an operation handler
func RegisterOperation(operationID string, handler OperationHandler)

// GetOperation returns a registered operation handler
func GetOperation(operationID string) (OperationHandler, bool)
```

**HTTP Client Methods:**

```go
// Call makes an HTTP request
func (c *HTTPClient) Call(ctx context.Context, req *Request) (*Response, error)

// SetAuthToken sets the authorization header
func (c *HTTPClient) SetAuthToken(token string, scheme string)

// ClearAuthToken removes the authorization header
func (c *HTTPClient) ClearAuthToken()

// GetConfig returns the current configuration
func (c *HTTPClient) GetConfig() *ClientConfig
```

**Error Handling:**

```go
// WASMError represents a structured error
type WASMError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// Error codes
const (
    ErrCodeInvalidConfig     = "INVALID_CONFIG"
    ErrCodeRequestFailed     = "REQUEST_FAILED"
    ErrCodeSerializationFail = "SERIALIZATION_FAILED"
    ErrCodeNetworkError      = "NETWORK_ERROR"
    ErrCodeTimeout           = "TIMEOUT"
)
```

## Examples

### Petstore Example

The `examples/petstore/` directory contains a sample OpenAPI specification for testing:

**OpenAPI Spec Features:**
- Three operations: `createPet`, `getPetById`, `findPetsByStatus`
- Path parameters: `petId` (int64)
- Query parameters: `status` (string)
- Request body: `Pet` object
- Response: `Pet` object or array of `Pet`
- Schema with enum: `status` field with values `available`, `pending`, `sold`

**Generate and Run:**

```bash
# Generate SDK from petstore example
make dev-generate

# Output will be in examples/petstore/generated/
# ├── generated.go    (Go client code)
# ├── sdk.ts          (TypeScript SDK)
# └── index.html      (Interactive demo page)
```

**Generated Code Preview:**

```go
// generated.go - Go client code
type Pet struct {
    ID     int64  `json:"id,omitempty"`
    Name   string `json:"name"`
    Status string `json:"status,omitempty"`
}

type CreatePetRequest struct {
    Body *Pet `json:"body,omitempty"`
}

func CreatePetRequestToRequest(params CreatePetRequest) runtime.Request {
    return runtime.Request{
        Method: "POST",
        Path:   "/pet",
        Body:   params.Body,
    }
}
```

```typescript
// sdk.ts - TypeScript SDK
export interface Pet {
    id?: number;
    name: string;
    status?: 'available' | 'pending' | 'sold';
}

export async function createPet(params: CreatePetRequest): Promise<HTTPResponse> {
    const request: HTTPRequest = {
        method: 'POST',
        path: '/pet',
        body: params.body,
    };
    return (window as any).wasmCallAPI('createPet', request);
}
```

**Testing the Demo:**

1. Build the WASM binary: `make build`
2. Generate the SDK: `make dev-generate`
3. Serve the `examples/petstore/generated/` directory
4. Open `index.html` in a browser
5. Click "Load WASM" to load the WebAssembly module
6. Enter the base URL and click "Initialize"
7. Test the API operations

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/norunners/vert` | v0.0.0-20221203075838-106a353d42dd | Go ↔ JavaScript value conversion for WASM |
| `gopkg.in/yaml.v3` | v3.0.1 | OpenAPI YAML specification parsing |

**Standard library packages used:**

| Package | Purpose |
|---------|---------|
| `net/http` | HTTP client implementation |
| `encoding/json` | JSON serialization/deserialization |
| `syscall/js` | JavaScript interop for WASM |
| `text/template` | Code generation templates |
| `reflect` | Runtime type inspection for conversion |
| `sync` | Thread-safe state management (RWMutex) |

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Developer Workflow                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     OpenAPI 3.x Specification                   │
│                     (YAML/JSON format)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Generator                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  OpenAPI     │  │   Model      │  │  Template    │         │
│  │  Parser      │→ │   Builder    │→ │  Renderer    │         │
│  │  (openapi.go)│  │(generator.go)│  │ (templates)  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  generated.go   │ │    sdk.ts       │ │   index.html    │
│  (Go Client)    │ │  (TypeScript)   │ │  (Demo Page)    │
└─────────────────┘ └─────────────────┘ └─────────────────┘
              │               │
              ▼               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      WASM Runtime                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  HTTP Client │  │   Exports    │  │  Converter   │         │
│  │  (client.go) │← │  (exports.go)│← │ (converter.go)│         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│         │                  ↑                   │                │
│         │                  │                   │                │
│         ▼                  ▼                   ▼                │
│  ┌─────────────────────────────────────────────────────┐      │
│  │              JavaScript Browser                     │      │
│  │  - WebAssembly.instantiateStreaming()               │      │
│  │  - wasmInitClient()                                 │      │
│  │  - wasmCallAPI()                                    │      │
│  │  - Response conversion via vert library             │      │
│  └─────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Generator Flow:
   OpenAPI YAML → Parse → Build Model → Render Templates → Generate Files
                                                                  │
   ┌──────────────────────────────────────────────────────────────┘
   │
   ├──→ generated.go (Go structs, request/response handlers, init())
   ├──→ sdk.ts (TypeScript interfaces, WASMSDK class, typed functions)
   └──→ index.html (Interactive demo with Tailwind CSS)

2. Runtime Flow:
   Browser → Load WASM → Initialize Client → Make API Calls
                           │                      │
                           ▼                      ▼
                    ┌─────────────┐        ┌─────────────┐
                    │  Config     │        │  HTTP       │
                    │  Management │        │  Client     │
                    └─────────────┘        └─────────────┘
                           │                      │
                           ▼                      ▼
                    ┌─────────────┐        ┌─────────────┐
                    │  Auth       │        │  Request    │
                    │  Tokens     │        │  Building   │
                    └─────────────┘        └─────────────┘
                           │                      │
                           └──────────┬───────────┘
                                      ▼
                              ┌─────────────┐
                              │  Go Runtime │
                              │  (WASM)     │
                              └─────────────┘
                                      │
                                      ▼
                              ┌─────────────┐
                              │  vert       │
                              │  Converter  │
                              └─────────────┘
                                      │
                                      ▼
                              ┌─────────────┐
                              │  JavaScript │
                              │  Response   │
                              └─────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run verification with `make verify`
5. Submit a pull request

## License

This project is licensed under the MIT License.
