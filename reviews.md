# 项目代码评审报告

## 项目概述

**项目名称**: go_wasm_lib2 - Generic WASM HTTP SDK Generator  
**模块路径**: github.com/fred29910/gowasm  
**Go版本**: 1.25.1  
**主要依赖**: 
- github.com/norunners/vert (Go-WASM 反射转换库)
- gopkg.in/yaml.v3 (YAML 解析)

---

## 架构设计评审

### ✅ 优点

1. **清晰的分层架构**
   - `cmd/generator` - 代码生成器 CLI
   - `cmd/runtime` - WASM 运行时入口
   - `pkg/generator` - 核心生成逻辑（OpenAPI 解析、模型构建、模板渲染）
   - `pkg/runtime` - WASM 运行时核心（HTTP 客户端、JS 互操作、Promise 处理、类型转换）

2. **双语言生成能力**
   - 同时生成 Go 客户端代码（用于 WASM 端）和 TypeScript SDK（用于浏览器端）
   - 生成的代码包含类型定义、请求构建器、调用方法

3. **OpenAPI 3.x 完整支持**
   - 支持 schemas、operations、parameters、requestBody、responses
   - 自动处理 path/query/body 参数
   - 支持 $ref 引用解析

4. **WASM 运行时设计合理**
   - 基于 Promise 的异步模式，符合 JS 生态
   - 完整的错误处理体系（WASMError、错误码）
   - 使用 `vert` 库处理 Go↔JS 值转换，减少手工序列化代码

### ⚠️ 需改进项

1. **Generator 结构体职责过重** (`generator.go:11-38`)
   - 同时持有配置、模板路径、运行时导入路径等
   - 建议拆分为 `Config` 和 `Generator` 两个结构体

2. **模板内嵌在代码中** (`go_templates.go:42-117`, `ts_templates.go:52-270`)
   - 模板过长，维护困难
   - 建议迁移到独立文件或使用 `embed` (Go 1.16+)

3. **缺乏单元测试**
   - 整个项目无 `_test.go` 文件
   - 核心逻辑（类型转换、模型构建、OpenAPI 解析）无测试覆盖

4. **错误处理不一致**
   - 部分函数返回 `error`，部分 panic
   - `generator.go:193-234` `goType` 遇到未知类型静默返回 `interface{}`

---

## 代码质量评审

### pkg/generator

| 文件 | 行数 | 问题 |
|------|------|------|
| `generator.go` | 326 | 函数过长 (`buildModel` 115行)，建议拆分 |
| `openapi.go` | 194 | 解析逻辑较清晰，但缺乏对 OpenAPI 完整规范的支持 |
| `types.go` | 126 | 命名转换函数重复 (`sortedMapKeys` 与 `generator.go` 重复) |
| `go_templates.go` | 126 | 模板字符串内嵌，难以维护 |
| `ts_templates.go` | 271 | 同上，且 TS 模板生成的路径参数处理有 bug (第195行覆盖问题) |

**关键问题**:
- `ts_templates.go:195` - `pathParams` 在循环中被覆盖，只有最后一个参数生效
- `generator.go:272-278` `docSchemaRef` 与 `openapi.go:188-194` `SchemaRef` 功能重复

### pkg/runtime

| 文件 | 行数 | 评价 |
|------|------|------|
| `client.go` | 241 | HTTP 客户端实现完整，支持超时、认证、配置管理 |
| `exports.go` | 296 | JS 导出函数完整，但 `initClient` 中 goroutine 可能导致竞态 |
| `promise.go` | 94 | Promise 封装良好，错误传播正确 |
| `converter.go` | 249 | 依赖 `vert` 库，大幅简化转换逻辑 |
| `error.go` | 69 | 错误码定义清晰，实现 `error` 接口 |

**关键问题**:
- `exports.go:40-88` `initClient` 在 goroutine 中修改 `e.client` 和 `e.initialized`，无同步机制
- `converter.go:134-165` `interfaceToJSValue` 对 `map[string]interface{}` 断言可能 panic
- 缺乏 WASM 导入函数的实现 (`exports.go:134` 注释 "// Go WASM imports")

---

## 生成代码质量评审

### Go 生成代码 (`generated.go`)

**优点**:
- 类型安全的请求结构体
- 自动路径参数序列化 (`int64ToString` 等)
- 运行时操作注册机制

**问题**:
1. `CreatePetRequestToRequest` 未处理 path 参数 (模板 bug)
2. `PathParams` 字段重复定义 (结构体字段 + map 字段)
3. 缺乏响应体反序列化辅助方法

### TypeScript 生成代码 (`sdk.ts`)

**优点**:
- 完整的类型定义
- `WASMSDK` 类封装良好
- 类型安全的 API 方法

**问题**:
1. `createPet` 等函数未处理 path 参数 (模板 bug，第195行)
2. `wasmCallAPI` 返回 `Promise<HTTPResponse>` 但未处理错误类型
3. Demo HTML 中的 `call{{.ID}}` 函数未传递参数

---

## 构建与工程化评审

### ✅ 优点
- 提供 `Makefile` 和 `Taskfile.yml` 双构建系统
- 支持标准 Go 和 TinyGo 双编译器构建
- `generate` 目标支持参数化规格文件路径

### ⚠️ 需改进项
1. **无 CI/CD 配置** - 缺少 GitHub Actions / GitLab CI
2. **无版本管理** - 无 `version` 标签、无语义化版本
3. **依赖管理** - `go.sum` 在 `.gitignore` 中 (建议提交)
4. **文档缺失** - 无 README、无 API 文档、无使用指南

---

## 安全性评审

### 潜在风险
1. **XSS 风险** - `demoHTMLTemplate` 直接插入 `.BaseURL` 等用户可控数据，无转义
2. **路径遍历** - `ResolvePath` 使用 `url.PathEscape` 但未验证路径参数
3. **原型污染** - `converter.go` `JSValueToMap` 直接遍历 `Object.keys`，未过滤 `__proto__`

### 建议
- 对模板插入数据进行 HTML/JS 转义
- 添加路径参数白名单验证
- 在 `JSValueToMap` 中过滤危险键名

---

## 性能评审

### WASM 体积
- 标准 Go 编译: ~2-5 MB (未压缩)
- TinyGo 编译: ~200-500 KB
- **建议**: 默认推荐 TinyGo，或提供压缩构建选项 (`-ldflags="-s -w"`)

### 运行时性能
- `vert` 库基于反射，性能一般
- 每次 API 调用创建新 goroutine (`exports.go:42, 96`)
- **建议**: 考虑 worker pool 复用 goroutine

---

## 总体评分

| 维度 | 评分 (1-5) | 说明 |
|------|------------|------|
| 架构设计 | 4 | 分层清晰，职责分离良好 |
| 代码质量 | 3 | 核心逻辑可读，但有重复代码和模板 bug |
| 测试覆盖 | 1 | **完全无测试**，高风险 |
| 文档完善 | 2 | 仅有代码注释，无用户文档 |
| 工程化 | 3 | 构建脚本完善，缺 CI/CD |
| 安全性 | 2 | 存在 XSS/原型污染风险 |
| **综合** | **2.5** | 原型可用，生产需大量补强 |

---

## 优先级建议

### 🔴 P0 (必须修复)
1. 修复 TS 模板 pathParams 覆盖 bug (`ts_templates.go:195`)
2. 修复 Go 模板 path 参数未处理 bug (`go_templates.go:74`)
3. 添加基础单元测试 (OpenAPI 解析、类型转换、命名转换)
4. 修复 `initClient` 竞态条件 (加锁或原子操作)

### 🟠 P1 (重要)
1. 将模板提取到独立文件，使用 `//go:embed`
2. 消除重复代码 (`docSchemaRef`/`SchemaRef`, `sortedMapKeys`)
3. 添加 CI/CD 流程 (GitHub Actions: build, test, generate)
4. 编写 README 和使用文档

### 🟡 P2 (改进)
1. 支持更多 OpenAPI 特性 (oneOf, anyOf, allOf, discriminator)
2. 生成的代码添加响应反序列化方法
3. WASM 体积优化 (默认 TinyGo, 添加压缩标志)
4. 安全加固 (模板转义、参数验证)

### 🟢 P3 (优化)
1. 实现 WASM 导入函数 (console.log, fetch polyfill 等)
2. 添加代码生成选项 (包名、模块名、运行时路径自定义)
3. 支持多文件 OpenAPI spec (`$ref` 外部文件)
4. 生成 OpenAPI 文档页面

---

## 附录: 文件清单

```
go_wasm_lib2/
├── cmd/
│   ├── generator/main.go       # 生成器 CLI 入口
│   └── runtime/main.go         # WASM 入口
├── pkg/
│   ├── generator/
│   │   ├── generator.go        # 核心生成逻辑 (326 行)
│   │   ├── openapi.go          # OpenAPI 解析 (194 行)
│   │   ├── types.go            # 类型定义 (126 行)
│   │   ├── go_templates.go     # Go 模板 (126 行)
│   │   └── ts_templates.go     # TS 模板 (271 行)
│   └── runtime/
│       ├── client.go           # HTTP 客户端 (241 行)
│       ├── exports.go          # JS 导出 (296 行)
│       ├── promise.go          # Promise 封装 (94 行)
│       ├── converter.go        # 类型转换 (249 行)
│       └── error.go            # 错误定义 (69 行)
├── examples/petstore/          # 示例项目
├── build/                      # 构建产物
├── Makefile / Taskfile.yml     # 构建脚本
└── go.mod / go.sum             # 依赖管理
```

---

*评审时间: 2026-06-14*  
*评审人: AI Code Reviewer*