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

### New in Project Optimization

- **TypeScript Enum Union Types** — OpenAPI enum fields generate proper union types (`'available' | 'pending' | 'sold'`) instead of generic `string`
- **Go Naming Conventions** — Generated code follows Go idioms (`ID`, `URL`, `HTTP`, `JSON`, `API`, `UUID`, `JWT`)
- **Multi-Response Support** — All documented HTTP response codes generate typed response structs/interfaces (including 4xx and 5xx)
- **Request Validation** — Optional `Validate() error` methods on request structs check required fields, enum values, and format constraints
- **Custom Templates** — Override default templates with `--go-template` and `--ts-template` CLI flags
- **CLI Enhancements** — `--dry-run` previews without generating files; `--verbose` shows detailed progress; `--output json` for CI/CD parsing

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
│   ├── main.go             #   app 装配 + ExitErrHandler (74 行)
│   ├── flags.go            #   flag 定义 (80 行)
│   ├── generate.go         #   runGenerate 核心逻辑 (132 行)
│   ├── init.go             #   runInit 逻辑 (22 行)
│   ├── wasm.go             #   runBuildWASM 逻辑 (31 行)
│   └── lint.go             #   runOxlint 逻辑 (58 行)
├── cmd/runtime/            # WASM 运行时入口
├── pkg/generator/          # 代码生成器
│   ├── generator.go        #   核心生成逻辑 (614 行)
│   ├── openapi.go          #   OpenAPI 解析 (408 行)
│   ├── types.go            #   类型定义 (200 行)
│   ├── go_templates.go     #   Go 模板渲染（预编译）(172 行)
│   ├── ts_templates.go     #   TS 模板渲染（预编译）(130 行)
│   └── templates/          #   内置模板文件
├── pkg/runtime/            # WASM 运行时
│   ├── runtime.go          #   Facade 层 (93 行)
│   ├── runtime_wasm.go     #   WASM 导出入口 (88 行)
│   ├── client/             #   HTTP 客户端 (351 行)
│   ├── errors/             #   错误处理 (138 行)
│   ├── validate/           #   验证函数 (118 行)
│   ├── convert/            #   类型转换（js/wasm）(302 行)
│   ├── wasm/               #   JS 导出（js/wasm）(353 行)
│   └── build/              #   构建工具 (164 行)
├── docs/                   # 文档目录
├── .github/workflows/      # CI/CD 配置
├── Makefile                # Make 构建脚本 (129 行)
└── Taskfile.yml            # Task 构建脚本 (195 行)
```

## Migration Guide: Naming Convention Changes

如果您之前已生成 SDK 代码，升级到当前版本后可能会遇到以下命名变更：

- `Id` → `ID` （Go 字段和类型名）
- `Url` → `URL`
- `Http` → `HTTP`
- `Json` → `JSON`
- `Api` → `API`
- `Uuid` → `UUID`
- `Jwt` → `JWT`

**迁移方法**：
1. 重新运行生成器：`make generate SPEC=path/to/openapi.yaml OUT=./output`
2. 对于现有的手写代码，可搜索并替换上述字段名
3. TypeScript 代码不受影响（仅增强枚举类型生成）

## License

[MIT](LICENSE)
