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
        Parser[OpenAPI 解析器<br/>openapi.go]
        ModelBuilder[模型构建器<br/>generator.go]
        GoRenderer[Go 模板渲染器<br/>go_templates.go]
        TSRenderer[TypeScript 模板渲染器<br/>ts_templates.go]
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
        Client[HTTP 客户端<br/>client.go]
        Exports[JS 导出函数<br/>exports.go]
        Converter[类型转换器<br/>converter.go]
        Promise[Promise 封装<br/>promise.go]
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

    Runtime --> Browser
    JS --> WASM
    WASM --> JS
```

## 核心组件说明

### 1. CLI 入口 (`cmd/generator/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `main.go` | 341 | CLI 应用入口，定义 `generate` 和 `init` 子命令 |

**支持的子命令：**

| 命令 | 用途 |
|------|------|
| `generate` | 从 OpenAPI 规范生成 SDK |
| `init` | 创建示例项目结构 |

### 2. WASM 运行时入口 (`cmd/runtime/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `main.go` | 9 | WASM 模块入口，调用 `runtime.ExportMain()` |

### 3. 代码生成器 (`pkg/generator/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `generator.go` | 563 | 核心生成逻辑：模型构建、编排 |
| `openapi.go` | 194 | OpenAPI 3.x 解析器 |
| `types.go` | 126 | 类型定义和命名转换 |
| `go_templates.go` | 126 | Go 模板渲染逻辑 |
| `ts_templates.go` | 271 | TypeScript 模板渲染逻辑 |

### 4. WASM 运行时核心 (`pkg/runtime/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `client.go` | 292 | HTTP 客户端实现 |
| `exports.go` | 361 | JavaScript 导出函数 |
| `promise.go` | 95 | Promise 封装 |
| `converter.go` | 296 | Go ↔ JS 类型转换 |
| `error.go` | 109 | 错误类型和错误码定义 |
| `build.go` | 138 | WASM 构建工具 |

## 数据流图

### 1. 代码生成流程

```mermaid
flowchart LR
    A[OpenAPI YAML] --> B[ParseOpenAPI]
    B --> C[Build Model<br/>生成 Schemas + Operations]
    C --> D{Render Templates}
    D --> E[generated.go]
    D --> F[go.mod]
    D --> G[main.go]
    D --> H[sdk.ts]
    D --> I[index.html]
    E --> J[BuildWASM]
    G --> J
    J --> K[main.wasm]
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
    W5 --> W6

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
    C --> D{有 pathParams?}
    D -->|是| E[解析 pathParams]
    D -->|否| F[构建 Request]
    E --> F
    F --> G[client.Call]
    G --> H[buildURL<br/>替换路径参数]
    H --> I[序列化 Body]
    I --> J[http.NewRequest]
    J --> K[设置 Headers]
    K --> L[http.Do]
    L --> M[读取 Response]
    M --> N[JSON 反序列化]
    N --> O[GoToJSValue]
    O --> P[返回 JS Promise]
```

## 模板系统架构

```mermaid
flowchart TB
    subgraph Templates["模板文件 (pkg/generator/templates/)"]
        T1[sdk.go.tmpl<br/>Go 客户端模板]
        T2[sdk.ts.tmpl<br/>TypeScript SDK 模板]
        T3[go.mod.tmpl<br/>Go 模块模板]
        T4[index.html.tmpl<br/>演示页面模板]
        T5[main.go.tmpl<br/>WASM 入口模板]
    end

    subgraph Data["模板数据模型"]
        D1[GenerationModel]
        D2[Config]
        D3[GeneratedSchema]
        D4[GeneratedOperation]
    end

    subgraph Functions["自定义模板函数"]
        F1[hasPrefix<br/>字符串前缀检查]
        F2[htmlEscape<br/>HTML 转义]
    end

    D1 --> Templates
    D2 --> Templates
    D3 --> Templates
    D4 --> Templates
    Functions --> Templates

    subgraph Output["生成结果"]
        O1[generated.go]
        O2[sdk.ts]
        O3[go.mod]
        O4[index.html]
        O5[main.go]
    end

    Templates --> Output
```

## 并发安全设计

```mermaid
flowchart LR
    subgraph Threads["并发场景"]
        T1[JS 主线程<br/>wasmSetAuthToken]
        T2[Go goroutine<br/>wasmCallAPI]
    end

    subgraph Lock["同步机制"]
        M[sync.RWMutex<br/>HTTPClient.mu]
    end

    subgraph Shared["共享状态"]
        S1[config.Headers<br/>map[string]string]
    end

    T1 -->|Lock| M
    T2 -->|RLock| M
    M --> S1
```

## 构建系统架构

```mermaid
flowchart TB
    subgraph BuildScripts["构建脚本"]
        Makefile[Makefile<br/>81 行]
        Taskfile[Taskfile.yml<br/>185 行]
    end

    subgraph Targets["构建目标"]
        Build[build<br/>标准 Go 编译]
        BuildTiny[build-tinygo<br/>TinyGo 编译]
        Generate[generate<br/>SDK 生成]
        Test[test<br/>单元测试]
        Verify[verify<br/>完整验证]
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
│   │   └── main.go              # CLI 入口 (341 行)
│   └── runtime/
│       └── main.go              # WASM 入口 (9 行)
├── pkg/
│   ├── generator/
│   │   ├── generator.go         # 核心生成逻辑 (563 行)
│   │   ├── openapi.go           # OpenAPI 解析 (194 行)
│   │   ├── types.go             # 类型定义 (126 行)
│   │   ├── go_templates.go      # Go 模板渲染 (126 行)
│   │   ├── ts_templates.go      # TS 模板渲染 (271 行)
│   │   └── templates/           # 模板文件
│   │       ├── sdk.go.tmpl
│   │       ├── sdk.ts.tmpl
│   │       ├── go.mod.tmpl
│   │       ├── index.html.tmpl
│   │       └── main.go.tmpl
│   └── runtime/
│       ├── client.go            # HTTP 客户端 (292 行)
│       ├── exports.go           # JS 导出 (361 行)
│       ├── promise.go           # Promise 封装 (95 行)
│       ├── converter.go         # 类型转换 (296 行)
│       ├── error.go             # 错误定义 (109 行)
│       └── build.go             # 构建工具 (138 行)
├── version/
│   └── version.go               # 版本信息
├── examples/
│   ├── petstore/
│   │   ├── openapi.yaml         # Petstore 示例规范
│   │   └── generated/           # 生成的 SDK
│   └── templates/               # 自定义模板示例
├── Makefile                     # Make 构建脚本
├── Taskfile.yml                 # Task 构建脚本
├── go.mod                       # Go 模块定义
└── package.json                 # npm 配置 (oxlint)
```
