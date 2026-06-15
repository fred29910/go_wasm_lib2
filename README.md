# go_wasm_lib2 - Generic WASM HTTP SDK Generator

A Go-based toolkit for building WebAssembly (WASM) HTTP clients and generating type-safe SDKs from OpenAPI 3.x specifications.

## Features

- **WASM Runtime** — HTTP client compiled to WebAssembly with Promise-based async API, Bearer token auth, and configurable timeout/headers/credentials
- **SDK Generator** — Generate Go client code, TypeScript SDK, and interactive demo HTML from OpenAPI 3.x specs
- **Dual Compiler Support** — Standard Go (2-5MB) or TinyGo (200-500KB) for smaller binaries
- **Type-Safe** — Auto-generate TypeScript interfaces and typed API functions from OpenAPI schemas
- **Secure** — Prototype pollution protection, path traversal prevention, structured error codes

## Quick Start

```bash
# 1. Build the WASM runtime
make build          # Standard Go (2-5MB)
make build-tinygo   # TinyGo (200-500KB, recommended)

# 2. Generate SDK from an OpenAPI spec
make generate SPEC=path/to/openapi.yaml OUT=./output

# 3. Or try the petstore example
make dev-generate
```

## Basic Usage

### Use Generated SDK in Browser

```typescript
import { WASMSDK, createPet, getPetById } from './generated/sdk';

const sdk = new WASMSDK('./main.wasm');
await sdk.load();
await sdk.init({ baseUrl: 'https://petstore3.swagger.io/api/v3' });

const response = await sdk.createPet({
  body: { name: 'Fluffy', status: 'available' }
});
```

### Direct WASM API Calls

```javascript
// Initialize
await window.wasmInitClient({
  baseUrl: 'https://api.example.com',
  timeout: 30,
  credentials: 'same-origin'
});

// Authenticate
window.wasmSetAuthToken('your-jwt-token');

// Call API
const response = await window.wasmCallAPI('getUsers', {
  method: 'GET',
  path: '/users',
  query: { page: '1', limit: '10' }
});
```

## Build Targets

| Target | Output | Size |
|--------|--------|------|
| `make build` | `build/main.wasm` | 2-5 MB |
| `make build-tinygo` | `build/tinymain.wasm` | 200-500 KB |
| `make generate` | `<out>/` | depends on compiler |

> Supports both `make` and `task` (Taskfile) build systems.

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System architecture with Mermaid diagrams |
| [docs/getting-started.md](docs/getting-started.md) | Setup guide and build system comparison |
| [docs/cli-reference.md](docs/cli-reference.md) | CLI command and flag reference |
| [docs/runtime-api.md](docs/runtime-api.md) | WASM exported functions and error codes |
| [docs/generator-api.md](docs/generator-api.md) | Template system and custom template variables |
| [docs/known-issues.md](docs/known-issues.md) | Known limitations and roadmap |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/norunners/vert` | Go ↔ JavaScript value conversion |
| `gopkg.in/yaml.v3` | OpenAPI YAML parsing |

## License

MIT
