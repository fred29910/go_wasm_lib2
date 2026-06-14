# go_wasm_lib2 - Generic WASM HTTP SDK Generator

A Go-based toolkit for building WebAssembly (WASM) HTTP clients and generating SDKs from OpenAPI specifications.

## Overview

This project provides two main components:

1. **WASM Runtime Library** - A JavaScript-compatible runtime library that can be compiled to WebAssembly for use in browser environments
2. **SDK Generator** - A code generator that creates HTTP client SDKs from OpenAPI specifications

## Features

- ✅ Build WebAssembly HTTP clients with standard Go compiler
- ✅ Build smaller WebAssembly binaries with TinyGo compiler
- ✅ Generate TypeScript/JavaScript SDKs from OpenAPI specs
- ✅ Support for HTTP/1.1 with promise-based API
- ✅ Automatic request/response conversion
- ✅ Type-safe API generation

## Quick Start

### Prerequisites

- Go 1.25.1 or later
- TinyGo (optional, for smaller builds)

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
# Generate SDK from your OpenAPI spec
make generate SPEC=path/to/openapi.yaml OUT=./output

# Example with petstore
make dev-generate
```

### Run Tests

```bash
# Test compilation
make test-compile

# Run Go tests
make test

# Full verification
make verify
```

## Usage

### Runtime Usage

The generated WASM binary (`build/main.wasm`) can be used in web applications:

```html
<script>
  const wasm = await WebAssembly.instantiateStreaming(fetch('build/main.wasm'));
  const client = wasm.exports;
  
  // Use the client to make HTTP requests
  const response = await client.fetch('https://api.example.com/data');
  const data = await client.read(response);
</script>
```

### Generated SDK Usage

After generating an SDK from an OpenAPI spec, you can use it to make HTTP requests:

```typescript
// Generated client code
const client = new PetstoreClient('https://petstore3.swagger.io/api/v3');

// Create a new pet
const pet = await client.createPet({
  name: 'Fluffy',
  status: 'available'
});

// Get pet by ID
const foundPet = await client.getPetById(123);
```

## Project Structure

```
.
├── cmd/
│   ├── runtime/          # WASM runtime entry point
│   └── generator/        # SDK generator CLI
├── pkg/
│   ├── runtime/          # Runtime library code
│   └── generator/        # Generator library code
├── examples/
│   └── petstore/         # Example OpenAPI spec
├── build/                # Generated WASM binaries
└── examples/*/generated/ # Generated SDKs
```

## Configuration

### OpenAPI Spec Requirements

The generator expects OpenAPI 3.0 specifications with:

- HTTP paths and methods
- Request/response schemas
- Parameter definitions

### Generated Code Options

```bash
# Run generator with custom options
make run:generator args="-spec=spec.yaml -out=./output -module=mydomain -package=client"
```

## Development

### Install TinyGo

```bash
# macOS
make install-tinygo-macos

# Linux
make install-tinygo-linux
```

### Development Workflow

```bash
# Full development setup
make dev

# Clean everything
make clean:all
```

## Examples

### Petstore Example

The `examples/petstore/` directory contains a sample OpenAPI specification that can be used to test the SDK generator:

```bash
# Generate SDK from petstore example
make dev-generate
```

This will generate a TypeScript SDK in `examples/petstore/generated/`.

## Testing

### Compilation Tests

```bash
# Test that the code compiles correctly
make test-compile
```

### Go Tests

```bash
# Run Go unit tests
make test
```

### Full Verification

```bash
# Run complete verification suite
make verify
```

This includes:
- Dependency management
- Build verification
- Compilation tests
- Example generation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests with `make verify`
5. Submit a pull request

## License

This project is licensed under the MIT License.
