# 架构设计文档

## 系统概述

`go_wasm_lib2` (模块路径: `github.com/fred29910/gowasm`) 是一个基于 Go 的 WebAssembly (WASM) HTTP SDK 生成器。它能够为 OpenAPI 3.x 规范自动生成类型安全的客户端 SDK，并编译为可在浏览器中运行的 WASM 模块。

## 系统架构图

```mermaid
flowchart TB
    subgraph Input["输入"]
        OpenAPI[OpenAPI 3.x 规范<br/>YAML/JSON]
    end

    subgraph Generator["代码生成器 (pkg/generator)"]
        Parser[OpenAPI 解析器<br/>openapi.go<br/>191 行]
        ModelBuilder[模型构建器<br/>generator.go<br/>582 行]
        GoRenderer[Go 模板渲染器<br/>go_templates.go<br/>102 行]
        TSRenderer[TypeScript 模板渲染器<br/>ts_templates.go<br/>114 行]
    end

    subgraph Output["生成产物"]
        GeneratedGo[generated.go<br/>Go 客户端代码]
        GeneratedMod[go.mod<br/>Go 模块定义]
        GeneratedMain[main.go<br/>WASM 入口]
        GeneratedTS[sdk.ts<br/>TypeScript SDK]
        GeneratedHTML[index.html<br/>演示页面]
    end

    subgraph Compiler["WASM 编译器"]
        TinyGo[TinyGo 编译器<br/>200-500KB]
        GoCompiler[Go 编译器<br/>2-5MB]
    end

    subgraph Runtime["WASM 运行时 (pkg/runtime)"]
        Client[HTTP 客户端<br/>client.go<br/>291 行]
        Exports[JS 导出函数<br/>exports.go<br/>360 行]
        Converter[类型转换器<br/>converter.go<br/>295 行]
        Promise[Promise 封装<br/>promise.go<br/>94 行]
        Error[错误处理<br/>error.go<br/>108 行]
        Build[构建工具<br/>build.go<br/>137 行]
    end

    subgraph Lint["代码检查"]
        Oxlint[oxlint<br/>TypeScript 代码质量]
    end

    subgraph Browser["浏览器环境"]
        JS[JavaScript 应用]
        WASM[WASM 模块]
    end

    OpenAPI --> Parser
    Parser --> ModelBuilder
    ModelBuilder --> GoRenderer
    ModelBuilder --> TSRenderer
    GoRenderer --> GeneratedGo
    GoRenderer --> GeneratedMod
    GoRenderer --> GeneratedMain
    TSRenderer --> GeneratedTS
    TSRenderer --> GeneratedHTML

    GeneratedGo --> Compiler
    GeneratedMain --> Compiler
    Compiler --> TinyGo
    Compiler --> GoCompiler

    TinyGo --> Runtime
    GoCompiler --> Runtime

    GeneratedTS --> Oxlint
    Runtime --> Browser
    JS --> WASM
    WASM --> JS
```

## 核心组件说明

### 1. CLI 入口 (`cmd/generator/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `main.go` | 348 | CLI 应用入口，定义 `generate`、`init`、`version` 子命令 |

**支持的子命令：**

| 命令 | 用途 |
|------|------|
| `generate` | 从 OpenAPI 规范生成 SDK |
| `init` | 创建示例项目结构 |
| `version` | 显示版本信息 |

### 2. WASM 运行时入口 (`cmd/runtime/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `main.go` | 8 | WASM 模块入口，调用 `runtime.ExportMain()` |

### 3. 代码生成器 (`pkg/generator/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `generator.go` | 582 | 核心生成逻辑：模型构建、编排 |
| `openapi.go` | 191 | OpenAPI 3.x 解析器 |
| `types.go` | 200 | 类型定义和命名转换（含缩写词表：ID, URL, HTTP...） |
| `go_templates.go` | 102 | Go 模板渲染逻辑 |
| `ts_templates.go` | 114 | TypeScript 模板渲染逻辑 |

### 4. WASM 运行时核心 (`pkg/runtime/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `client.go` | 295 | HTTP 客户端实现，支持路径参数、查询参数、请求体 |
| `exports.go` | 349 | JavaScript 导出函数（init/callAPI/auth/getConfig） |
| `converter.go` | 295 | Go ↔ JavaScript 类型转换（通过 `vert` 库），含原型污染防护 |
| `promise.go` | 94 | Promise 封装：CreatePromise / ResolvePromise / RejectPromise |
| `error.go` | 138 | 结构化错误类型，含错误码、文件位置、行号、修复建议，支持 `Unwrap` |
| `build.go` | 155 | WASM 构建工具（auto/tinygo/go 编译器选择） |
| `validator.go` | 174 | 共享验证函数（email/uuid/datetime/enum），从模板中抽离以减小 WASM 体积 |

### 5. 内置模板 (`pkg/generator/templates/`)

| 模板文件 | 行数 | 输出文件 | 用途 |
|----------|------|----------|------|
| `sdk.go.tmpl` | 167 | `generated.go` | Go 客户端代码：schema 结构体、请求/响应类型、验证方法、辅助函数 |
| `sdk.ts.tmpl` | 170 | `sdk.ts` | TypeScript SDK：接口定义、WASMSDK 类、类型化 API 函数 |
| `go.mod.tmpl` | 7 | `go.mod` | Go 模块定义 |
| `main.go.tmpl` | 11 | `main.go` | WASM 入口文件 |
| `index.html.tmpl` | 769 | `index.html` | 交互式演示页面（Tailwind CSS） |

### 6. 构建系统

| 文件 | 行数 | 职责 |
|------|------|------|
| `Makefile` | 107 | GNU Make 构建系统 |
| `Taskfile.yml` | 194 | Task runner 构建系统（跨平台） |

## 安全架构

### 原型污染防护

`converter.go` 在解析 JavaScript 对象时自动过滤危险键：

```go
var dangerousJSKeys = map[string]bool{
    "__proto__": true, "constructor": true, "prototype": true,
    "__defineGetter__": true, "__defineSetter__": true, "__lookupGetter__": true,
    "__lookupSetter__": true, "hasOwnProperty": true, "isPrototypeOf": true,
    "propertyIsEnumerable": true, "toLocaleString": true, "toString": true, "valueOf": true,
}
```

### 路径遍历防护

`client.go` 的 `ResolvePath` 使用正则表达式 `\{([^}]+)\}` 进行确定性的单次替换，避免了 map 遍历顺序随机导致的参数注入漏洞。`safePathParam` 检测并拒绝：
- 包含 `..` 的路径值
- 包含 `//` 的路径值
- 以 `/` 开头的路径值

### OOM 防护

`client.go` 的 `Call` 方法使用 `io.LimitReader(resp.Body, 10<<20)` 限制响应体最大读取 10 MB，防止恶意服务端返回超大报文导致 WASM 内存溢出。

### 并发安全

| 组件 | 保护机制 | 保护对象 |
|------|----------|----------|
| `HTTPClient` | `sync.RWMutex` (c.mu) | `config.Headers` 读写、初始化状态 |
| `operations` 注册表 | `sync.RWMutex` (operationsMu) | 操作处理函数的并发注册与查找 |
| `WASMExports` | `sync.RWMutex` (e.mu) | 客户端初始化状态的读写 |

```mermaid
flowchart LR
    subgraph Read["读操作 (RLock)"]
        R1[Call<br/>读取 Headers]
        R2[GetAuthToken<br/>读取配置]
    end

    subgraph Write["写操作 (Lock)"]
        W1[SetAuthToken<br/>写入 Header]
        W2[InitClient<br/>初始化状态]
    end

    subgraph Lock["sync.RWMutex"]
        M[c.mu / operationsMu]
    end

    R1 --> M
    R2 --> M
    W1 --> M
    W2 --> M
```

### 结构化错误处理

`error.go` 的 `WASMError` 包含丰富的上下文信息，并通过 `runtime.Caller` 自动捕获调用位置，支持 `Unwrap()` 以兼容 `errors.Is`/`errors.As`：

```go
type WASMError struct {
    Code       string `json:"code"`       // 错误码，如 "TIMEOUT"
    Message    string `json:"message"`    // 人类可读的错误描述
    Details    string `json:"details"`    // 底层错误详情
    FilePath   string `json:"filePath"`   // 出错源文件（自动捕获）
    LineNumber int    `json:"lineNumber"` // 出错行号（自动捕获）
    Suggestion string `json:"suggestion"` // 修复建议
}
```

## 数据流图

### 1. 代码生成流程

```mermaid
flowchart LR
    A[OpenAPI YAML] --> B[ParseOpenAPI<br/>openapi.go]
    B --> C[Build Model<br/>生成 Schemas + Operations<br/>generator.go]
    C --> D{Render Templates<br/>go_templates.go / ts_templates.go}
    D --> E[generated.go<br/>Go structs + validation]
    D --> F[go.mod<br/>模块定义]
    D --> G[main.go<br/>WASM 入口]
    D --> H[sdk.ts<br/>TypeScript SDK]
    D --> I[index.html<br/>演示页面]
    E --> J[BuildWASM<br/>build.go]
    G --> J
    J --> K[main.wasm]
    H --> L[oxlint<br/>代码检查]
```

### 2. WASM 运行时流程

```mermaid
flowchart TB
    subgraph JS["JavaScript 侧"]
        J1[加载 WASM]
        J2[wasmInitClient]
        J3[wasmCallAPI]
        J4[wasmSetAuthToken]
    end

    subgraph WASM["WASM 侧 (Go)"]
        W1[ExportMain]
        W2[initClient<br/>初始化 HTTP 客户端]
        W3[callAPI<br/>执行 HTTP 请求]
        W4[setAuthToken<br/>设置认证令牌]
        W5[HTTP Client]
        W6[vert Converter]
    end

    J1 --> W1
    J2 --> W2
    J3 --> W3
    J4 --> W4

    W2 --> W5
    W3 --> W5
    W3 --> W6
    W4 --> W5

    subgraph Network["网络"]
        N1[HTTP/HTTPS]
    end

    W5 --> N1
    N1 --> W5
```

### 3. 请求处理流程

```mermaid
flowchart LR
    A[JS 调用<br/>wasmCallAPI] --> B[解析 operationId]
    B --> C[解析 request 对象]
    C --> D[解析 pathParams → map]
    C --> E[解析 query → url.Values]
    C --> F[解析 headers → map]
    C --> G[解析 body → interface{}]
    D --> H[构建 Request 结构体]
    E --> H
    F --> H
    G --> H
    H --> I[client.Call]
    I --> J[buildURL<br/>url.JoinPath 拼接基础 URL + 正则替换路径参数]
    J --> K[序列化 Body]
    K --> L[http.NewRequest]
    M --> N[http.Do]
    N --> O[读取 Response<br/>LimitReader 10MB 上限]
    O --> P[JSON 反序列化]
    P --> Q[GoToJSValue]
    Q --> R[返回 JS Promise]
```

## 构建系统架构

```mermaid
flowchart TB
    subgraph BuildScripts["构建脚本"]
        Makefile[Makefile<br/>107 行]
        Taskfile[Taskfile.yml<br/>194 行]
    end

    subgraph Targets["构建目标"]
        Build[build<br/>标准 Go 编译]
        BuildTiny[build-tinygo<br/>TinyGo 编译]
        Generate[generate<br/>SDK 生成]
        Test[test<br/>单元测试]
        Verify[verify<br/>完整验证]
        LintTs[lint-ts<br/>TS 代码检查]
    end

    subgraph Compilers["编译器选择"]
        CompilerAuto[auto<br/>自动检测]
        CompilerTiny[tinygo<br/>TinyGo]
        CompilerGo[go<br/>标准 Go]
    end

    Makefile --> Targets
    Taskfile --> Targets
    Targets --> Compilers
```

## 文件清单

```
go_wasm_lib2/
├── cmd/
│   ├── generator/
│   │   └── main.go              # CLI 入口 (348 行)
│   └── runtime/
│       └── main.go              # WASM 入口 (8 行)
├── pkg/
│   ├── generator/
│   │   ├── generator.go         # 核心生成逻辑 (582 行)
│   │   ├── openapi.go           # OpenAPI 解析 (191 行)
│   │   ├── types.go             # 类型定义 (200 行)
│   │   ├── go_templates.go      # Go 模板渲染 (102 行)
│   │   ├── ts_templates.go      # TS 模板渲染 (114 行)
│   │   ├── generator_test.go    # 生成器测试 (1718 行)
│   │   ├── openapi_test.go      # 解析器测试 (626 行)
│   │   └── types_test.go        # 类型测试 (193 行)
│   │   └── templates/           # 模板文件
│   │       ├── sdk.go.tmpl      # Go 客户端模板 (167 行)
│   │       ├── sdk.ts.tmpl      # TypeScript SDK 模板 (170 行)
│   │       ├── go.mod.tmpl       # Go 模块模板 (7 行)
│   │       ├── index.html.tmpl  # 演示页面模板 (769 行)
│   │       └── main.go.tmpl     # WASM 入口模板 (11 行)
│   └── runtime/
│       ├── client.go            # HTTP 客户端 (295 行)
│       ├── exports.go           # JS 导出 (349 行)
│       ├── promise.go           # Promise 封装 (94 行)
│       ├── converter.go         # 类型转换 (295 行)
│       ├── error.go             # 错误定义 (138 行)
│       ├── build.go             # 构建工具 (155 行)
│       ├── validator.go         # 共享验证函数 (174 行)
│       ├── build_test.go        # 构建测试 (152 行)
│       └── client_test.go       # 客户端测试 (229 行)
├── version/
│   └── version.go               # 版本信息 (11 行)
├── examples/
│   ├── petstore/
│   │   └── openapi.yaml         # Petstore 示例规范
│   └── templates/               # 自定义模板示例
├── Makefile                     # Make 构建脚本 (107 行)
├── Taskfile.yml                 # Task 构建脚本 (194 行)
├── go.mod                       # Go 模块定义
├── package.json                 # npm 配置 (oxlint)
└── oxlintrc.json                # oxlint 配置
```
