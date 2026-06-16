# 已知问题和限制

本文档记录了项目中已知的 Bug、限制和待改进项。

## 🔴 严重问题 (P0)

> ✅ 所有 P0 问题已在当前版本中修复。

### 1. 模板中 pathParams 覆盖问题 ✅ 已修复

**文件**: `pkg/generator/templates/sdk.ts.tmpl`

**状态**: ✅ 已修复 — 当前模板正确生成 pathParams 映射。

### 2. Go 模板缺少 stringToString 函数 ✅ 已修复

**文件**: `pkg/generator/templates/sdk.go.tmpl`

**状态**: ✅ 已修复 — 模板已包含 `stringToString` 辅助函数。

### 3. Makefile dev-generate 命令格式过时 ✅ 已修复

**文件**: `Makefile`

**状态**: ✅ 已修复 — Makefile 已更新为 `gowasm-generator generate -s ...`。

### 4. 并发数据竞争引发 Panic ✅ 已修复

**文件**: `pkg/runtime/client.go`

**问题**: `config.Headers` 的读写未加锁，`wasmCallAPI` 在 goroutine 中读取 Headers 时，JS 主线程调用 `wasmSetAuthToken` 写入 Map，导致 `Concurrent map iteration and map write` Panic。

**状态**: ✅ 已修复 — `HTTPClient` 新增 `sync.RWMutex`（`c.mu`），`Call()` 使用 `RLock` 读取 Headers，`SetAuthToken()`/`ClearAuthToken()` 使用 `Lock` 写入。

### 5. 路径参数插值完全瘫痪 ✅ 已修复

**文件**: `pkg/runtime/exports.go`, `pkg/runtime/client.go`

**问题**: `exports.go` 未解析 `pathParams`，`buildURL` 硬编码传 `nil` 给 `ResolvePath`。

**状态**: ✅ 已修复 — `exports.go` 现在解析 `pathParams` 并传递给 `Request.PathParams`，`client.Call()` 将其传给 `buildURL` → `ResolvePath`。

## 🟠 重要问题 (P1)

### 1. 生成的 Go 代码与 WASM 运行时未完全集成

**描述**: 生成的 `generated.go` 包含完整的类型定义和验证方法，但当前 WASM 运行时 (`exports.go` 的 `callAPI`) 直接忽略 `operationId`（代码注释 `_ = args[0].String() // operationId for future use`），使用通用的 `client.Call` 兜底发送。

**影响**: 生成的类型安全验证逻辑在浏览器中不生效。

**建议方案**:
- **方案 A (通用客户端)**: 若定位是通用 HTTP 调用，移除 README 中关于 Go Type-safe 特性的说明，移除 `generated.go` 的生成逻辑，仅保留 `sdk.ts`。
- **方案 B (强类型拦截)**: 若要实现完整的操作路由验证，需要让生成器输出 `main.go`（引入 `generated.go` 和 `runtime`），修改 `exports.go` 使 `callAPI` 通过 `operationID` 去 `GetOperation(opId)` 并执行相应的回调。

### 2. 查询参数不支持多值数组

**文件**: `pkg/runtime/client.go`

**问题**: Request 结构体中 `Query` 字段在 JavaScript 接口中被约束为 `Record<string, string>`，不支持同名多值的数组查询参数（如 `?status=available&status=pending`）。

**建议**: 重构为 `url.Values` 类型支持多值。

### 3. 请求体仅支持 JSON 序列化

**文件**: `pkg/runtime/client.go`

**问题**: 所有请求体强制使用 JSON 序列化，不支持 `FormData`、`ArrayBuffer` 等其他格式。

**建议**: 允许透传 `ArrayBuffer` 或 `Blob` 类型。

## 🟡 改进项 (P2)

### 1. 模板系统与运行时架构割裂

**描述**: 当前系统分为 `runtime` 与 `generator` 两部分。生成器输出了 `generated.go` 代码，但官方默认编译为 WASM 的入口是 `cmd/runtime/main.go`，该入口仅导入泛用的底层客户端 (`ExportMain`)，完全没有导入用户生成的代码。

**相关文档**: `docs/reviews/project_review.md`

### 2. 单元测试覆盖不完整

**描述**: `pkg/generator/`（`generator_test.go`、`openapi_test.go`、`types_test.go`）和 `pkg/runtime/`（`client_test.go`、`build_test.go`）已包含单元测试。尚未覆盖的文件包括：`converter.go`、`promise.go`、`exports.go`、`validator.go`。

**风险**: 中风险 — 核心逻辑已有测试覆盖，剩余未覆盖文件风险较低。

### 3. 不支持 OpenAPI 高级特性

**不支持的特性**:
- `oneOf` / `anyOf` / `allOf` — 组合 schema
- `discriminator` — 多态类型
- 外部 `$ref` 引用 — 仅支持内部引用

### 4. 错误处理不一致

**描述**:
- 部分函数返回 `error`
- 部分函数 panic
- `generator.go:193-234` 的 `goType` 遇到未知类型静默返回 `interface{}`

## 🟢 优化项 (P3)

### 1. 安全加固

- **XSS 风险**: 演示 HTML 模板直接插入 `.BaseURL` 等用户可控数据
- **路径参数静默失败**: `ResolvePath` 检测到非法输入返回 `""`，建议改为返回错误

### 2. WASM 体积优化

| 编译器 | 输出大小 | 建议 |
|--------|----------|------|
| 标准 Go | 2-5 MB | 添加 `-ldflags="-s -w"` 压缩选项 |
| TinyGo | 200-500 KB | 默认推荐 |

### 3. CI/CD 配置

**缺失**:
- GitHub Actions / GitLab CI 配置
- 自动化测试、构建、发布流程

### 4. 版本管理

**缺失**:
- 语义化版本标签
- `go.sum` 在 `.gitignore` 中（建议提交）

## API 兼容性说明

### 已知限制

1. **认证**: 仅支持 Bearer Token，不支持 OAuth2、API Key 等其他认证方式
2. **重试**: 无自动重试机制
3. **拦截器**: 不支持请求/响应拦截器
4. **缓存**: 无内置缓存机制
5. **SSE**: 不支持 Server-Sent Events
6. **WebSocket**: 不支持 WebSocket 协议

### TypeScript SDK 已知问题

1. **WASMSDK.load() 实现**: 当前实现使用 `WebAssembly.instantiate()`，与运行时实际加载方式不匹配
2. **错误类型**: `wasmCallAPI` 返回 `Promise<HTTPResponse>` 但未处理 WASMError 类型

## 代码评审报告摘要

完整评审报告见 `docs/reviews/project_review.md`

| 维度 | 评分 (1-5) | 说明 |
|------|------------|------|
| 架构设计 | 4 | 分层清晰，职责分离良好 |
| 代码质量 | 3 | 核心逻辑可读，但有重复代码 |
| 测试覆盖 | 2 | 核心包已有测试覆盖，部分文件未覆盖 |
| 文档完善 | 4 | 完整 docs/ 文档体系，含评审报告 |
| 工程化 | 3 | 构建脚本完善，缺 CI/CD |
| 安全性 | 4 | 已修复并发和路径遍历风险，XSS 已处理 |
| **综合** | **3.0** | 原型可用，生产需大量补强 |

## 修改历史

| 日期 | 修改内容 |
|------|----------|
| 2026-06-16 | 更新文档：修正测试覆盖描述、更新评分、消除模板代码重复 |
| 2026-06-15 | 更新文档评分：文档完善 2→3，安全性 2→3，综合 2.5→2.8 |
| 2026-06-15 | 创建完整 docs/ 文档目录 |
| 2026-06-14 | 修复 pathParams 覆盖问题 |
| 2026-06-14 | 添加 stringToString 函数 |
| 2026-06-14 | 更新 Makefile dev-generate 命令 |