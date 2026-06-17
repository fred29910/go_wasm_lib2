# go_wasm_lib2 - Generic WASM HTTP SDK Generator

A Go-based toolkit for building WebAssembly (WASM) HTTP clients and generating type-safe SDKs from OpenAPI 3.x specifications.

## Features

- **WASM Runtime** — HTTP client compiled to WebAssembly with Promise-based async API, Bearer token auth, and configurable timeout/headers/credentials
- **SDK Generator** — Generate Go client code, TypeScript SDK, and interactive demo HTML from OpenAPI 3.x specs
- **Dual Compiler Support** — Standard Go (2-5MB) or TinyGo (200-500KB) for smaller binaries
- **Type-Safe** — Auto-generate TypeScript interfaces and typed API functions from OpenAPI schemas
- **Secure** — Prototype pollution protection, path traversal prevention, structured error codes
- **Structured Logging** — `log/slog` based logging with key-value output for machine readability
- **Tested** — 6 test packages with race condition detection (`make test-race`)

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
| `make test` | — | Run all tests |
| `make test-race` | — | Run tests with race detection |
| `make verify` | — | Full project verification |

> Supports both `make` and `task` (Taskfile) build systems.

## CLI Usage

```bash
gowasm-generator generate \
  -s ./api/openapi.yaml \        # OpenAPI spec (required)
  -o ./output \                  # Output directory (default: ./generated)
  -m github.com/myorg/myproject \ # Go module name
  -p apiclient \                 # Go package name
  --compiler tinygo \            # Compiler: auto|tinygo|go (default: auto)
  --wasm \                       # Build WASM after generation (default: false)
  --wasm-out ./build/sdk.wasm \  # WASM output path
  --validation \                 # Generate validation methods (default: true)
  -V, --verbose \                # Detailed progress
  --dry-run                      # Preview mode
```

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System architecture with Mermaid diagrams |
| [docs/getting-started.md](docs/getting-started.md) | Setup guide and build system comparison |
| [docs/cli-reference.md](docs/cli-reference.md) | CLI command and flag reference |
| [docs/runtime-api.md](docs/runtime-api.md) | WASM exported functions and error codes |
| [docs/generator-api.md](docs/generator-api.md) | Template system and custom template variables |
| [docs/known-issues.md](docs/known-issues.md) | Known limitations and roadmap |
| [CHANGELOG.md](CHANGELOG.md) | Version history and changes |
| [docs/reviews/comprehensive-review-v2-2026-06-17.md](docs/reviews/comprehensive-review-v2-2026-06-17.md) | Comprehensive code review report |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/norunners/vert` | Go ↔ JavaScript value conversion |
| `gopkg.in/yaml.v3` | OpenAPI YAML parsing |
| `github.com/urfave/cli/v2` | CLI framework |

## Testing

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run specific package tests
go test -v ./pkg/generator/
go test -v ./pkg/runtime/client/
go test -v ./pkg/runtime/errors/
go test -v ./pkg/runtime/validate/
```

## Project Structure

```

go_wasm_lib2/
├── cmd/generator/          # CLI 入口（6 个文件）
│   ├── main.go             #   app 装配 + ExitErrHandler
│   ├── flags.go            #   flag 定义
│   ├── generate.go         #   runGenerate 核心逻辑
│   ├── init.go             #   runInit 逻辑
│   ├── wasm.go             #   runBuildWASM 逻辑
│   └── lint.go             #   runOxlint 逻辑
├── cmd/runtime/            # WASM 运行时入口
├── pkg/generator/          # 代码生成器
│   ├── generator.go        #   核心生成逻辑
│   ├── openapi.go          #   OpenAPI 解析
│   ├── types.go            #   类型定义
│   ├── go_templates.go     #   Go 模板渲染（预编译）
│   ├── ts_templates.go     #   TS 模板渲染（预编译）
│   └── templates/          #   内置模板文件
├── pkg/runtime/            # WASM 运行时
│   ├── runtime.go          #   Facade 层
│   ├── client/             #   HTTP 客户端
│   ├── errors/             #   错误处理
│   ├── validate/           #   验证函数
│   ├── convert/            #   类型转换（js/wasm）
│   ├── wasm/               #   JS 导出（js/wasm）
│   └── build/              #   构建工具
├── docs/                   # 文档目录
├── .github/workflows/      # CI/CD 配置
├── Makefile                # Make 构建脚本
└── Taskfile.yml            # Task 构建脚本
```

## License

[MIT](LICENSE)
