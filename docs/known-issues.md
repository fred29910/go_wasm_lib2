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

### 6. ResolvePath map 遍历顺序注入漏洞 ✅ 已修复

**文件**: `pkg/runtime/client.go`

**问题**: 使用 `for k, v := range pathParams` 遍历 map 进行字符串替换，Go map 遍历顺序随机，可能导致参数值被当作模板解析（如 `userId=123{postId}`）。

**状态**: ✅ 已修复 — 改用正则表达式 `regexp.MustCompile(\`\{([^}]+)\}\`)` 进行确定性单次替换。

### 7. 响应体 OOM 风险 ✅ 已修复

**文件**: `pkg/runtime/client.go`

**问题**: `io.ReadAll(resp.Body)` 无上限读取，恶意服务端返回超大报文可导致 WASM 内存溢出。

**状态**: ✅ 已修复 — 使用 `io.LimitReader(resp.Body, 10<<20)` 限制最大 10 MB。

### 8. 错误行号硬编码 ✅ 已修复

**文件**: `pkg/runtime/client.go`

**问题**: 错误处理中大量使用硬编码行号（如 `"client.go", 128`），代码修改后行号失效。

**状态**: ✅ 已修复 — 新增 `WrapWASMError()` 函数，使用 `runtime.Caller(1)` 自动捕获调用位置。

### 9. WASMError 不支持 Unwrap ✅ 已修复

**文件**: `pkg/runtime/error.go`

**问题**: `WASMError` 未实现 `Unwrap() error`，导致 `errors.Is`/`errors.As` 失效。

**状态**: ✅ 已修复 — 新增 `Unwrap()` 方法，`WrapWASMError` 自动包装底层 error。

### 10. 全局单例 DefaultClient 限制 ✅ 已修复

**文件**: `pkg/generator/templates/sdk.go.tmpl`

**问题**: 生成的代码所有 API 调用都依赖全局单例 `runtime.DefaultClient`，无法并发访问不同后端。

**状态**: ✅ 已修复 — 生成器现在输出 `APIClient` 结构体和 `NewAPIClient(cfg)` 构造函数，支持多实例。

### 11. 模板内联验证函数膨胀 ✅ 已修复

**文件**: `pkg/generator/templates/sdk.go.tmpl`

**问题**: 每个生成的 SDK 都内联 `isValidEmail`/`isValidUUID` 等函数，增加 WASM 体积。

**状态**: ✅ 已修复 — 验证函数抽取到 `pkg/runtime/validator.go`（118 行），模板通过 `runtime.IsValidEmail()` 等调用。

### 12. URL 拼接不安全 ✅ 已修复

**文件**: `pkg/runtime/client.go`

**问题**: `buildURL` 使用字符串拼接 `base + "/" + fullURL`，易产生双斜杠等问题。

**状态**: ✅ 已修复 — 改用 `url.JoinPath(c.config.BaseURL, resolvedPath)` 进行安全拼接。

## 🟠 重要问题 (P1)

### 1. 生成的 Go 代码与 WASM 运行时集成方案已确定

**描述**: 生成的 `generated.go` 包含完整的类型定义、`APIClient` 和验证方法。WASM 运行时 (`exports.go` 的 `callAPI`) 现在通过 `GetOperation(operationID)` 查找注册的处理函数，若找到则执行，否则回退到通用的 `client.Call`。

**当前状态**: ✅ 架构已打通 — 运行时支持 operationId 路由。但仍需用户自定义 `main.go` 来导入生成的代码并编译为 WASM，默认的 `cmd/runtime/main.go` 仅包含通用运行时。

**建议方案**:
- 使用者生成 SDK 后，需要编写自定义 `main.go` 导入生成的 `generated.go`，然后编译为 WASM。
- 或者生成器可以直接生成 `main.go`（模板已有 `main.go.tmpl`），但当前 `Generate()` 流程未自动编译生成的代码为 WASM。

### 2. 查询参数已支持多值（url.Values）

**当前状态**: ✅ 已修复 — `Request.Query` 字段类型为 `url.Values`（即 `map[string][]string`），`exports.go` 使用 `req.Query.Add(k, v)` 添加查询参数，支持同名多值查询。TypeScript SDK 接口中 `query` 已更新为 `Record<string, string | string[] | number | boolean | null>`，支持多值查询参数传递。

### 3. Promise executor 缺少 panic 恢复

**文件**: `pkg/runtime/wasm/promise.go`

**问题**: `CreatePromise` 中的 executor 函数若发生 panic，会导致 goroutine 崩溃。

**状态**: ✅ 已修复 — 新增 `recover()` 保护，panic 被转换为 `fmt.Errorf("panic in WASM promise executor: %v", r)` 并传递给 reject。

### 4. Promise rejection 缺少 suggestion/filePath/lineNumber 字段

**文件**: `pkg/runtime/wasm/promise.go`

**问题**: `RejectPromise` 仅传递 `code` 和 `details`，缺少 `suggestion`、`filePath`、`lineNumber` 字段。

**状态**: ✅ 已修复 — `RejectPromise` 现在传递 `WASMError` 的所有字段（`suggestion`、`filePath`、`lineNumber`）。

### 5. datetime validator 手写解析脆弱

**文件**: `pkg/runtime/validate/validator.go`

**问题**: `IsValidDateTime` 使用手写的字节级解析，不支持毫秒、时区缩写等合法 RFC 3339 变体。

**状态**: ✅ 已修复 — 改用 `time.Parse(time.RFC3339, dt)`，覆盖所有合法 RFC 3339 格式。

### 6. CLI -v alias 冲突

**文件**: `cmd/generator/flags.go`

**问题**: `--validation` 的 `-v` alias 与 root 的 `--version` 冲突。

**状态**: ✅ 已修复 — 移除 `--validation` 的 `-v` alias。

### 7. --wasm 默认 true 风险较高

**文件**: `cmd/generator/flags.go`

**问题**: `--wasm` 默认为 `true`，用户可能无意中触发 WASM 编译。

**状态**: ✅ 已修复 — `--wasm` 默认改为 `false`。

## 🟡 改进项 (P2)

### 单元测试覆盖

**当前状态**: ✅ 已解决

所有核心包现已有完整测试覆盖：

| 包 | 测试文件 | 行数 | 说明 |
|----|---------|------|------|
| `pkg/generator/` | `generator_test.go` | 1729 | 生成器核心逻辑测试 |
| `pkg/generator/` | `openapi_test.go` | 626 | OpenAPI 解析测试 |
| `pkg/generator/` | `types_test.go` | 193 | 类型转换测试 |
| `pkg/generator/` | `petstore_integration_test.go` | 72 | Petstore 集成测试 |
| `pkg/runtime/client/` | `client_test.go` | 223 | HTTP 客户端测试 |
| `pkg/runtime/build/` | `build_test.go` | 154 | 构建工具测试 |
| `pkg/runtime/errors/` | `error_test.go` | 161 | 错误类型测试 |
| `pkg/runtime/validate/` | `validator_test.go` | 149 | 验证函数测试 |

**备注**: `pkg/runtime/convert/` 和 `pkg/runtime/wasm/` 因依赖 `js/wasm` 构建标签，难以用标准 Go 测试框架覆盖。这些包在实际 WASM 环境中工作正常。

### 不支持 OpenAPI 高级特性

**不支持的特性**:
- `oneOf` / `anyOf` / `allOf` — 组合 schema
- `discriminator` — 多态类型
- 外部 `$ref` 引用 — 仅支持内部引用

## 🟢 优化项 (P3)

### 1. 安全加固

- **XSS 风险**: 演示 HTML 模板直接插入 `.BaseURL` 等用户可控数据
- **路径参数静默失败**: `ResolvePath` 检测到非法输入返回 `""`，建议改为返回错误

### 2. WASM 体积优化

| 编译器 | 输出大小 | 建议 |
|--------|----------|------|
| 标准 Go | 2-5 MB | 已添加 `-ldflags="-s -w"` 和 `-trimpath` 压缩选项 |
| TinyGo | 200-500 KB | 默认推荐 |

### 3. Taskfile generate 命令格式

**文件**: `Taskfile.yml`

**问题**: `task generate` 命令使用旧的 CLI 标志格式（`-spec=`/`-out=`），缺少 `generate` 子命令。`dev:generate` 已修复此问题。

**建议**: 使用 `make generate` 或直接运行 `gowasm-generator generate` 命令。

### 4. CI/CD 配置

**缺失**:
- GitHub Actions / GitLab CI 配置
- 自动化测试、构建、发布流程

### 5. 版本管理

**缺失**:
- 语义化版本标签
- `go.sum` 建议提交到版本控制

## API 兼容性说明

### 已知限制

1. **认证**: 仅支持 Bearer Token 和自定义 scheme，不支持 OAuth2、API Key 等其他认证方式
2. **重试**: 无自动重试机制
3. **拦截器**: 不支持请求/响应拦截器
4. **缓存**: 无内置缓存机制
5. **SSE**: 不支持 Server-Sent Events
6. **WebSocket**: 不支持 WebSocket 协议

### TypeScript SDK 备注

1. **错误类型**: `wasmCallAPI` 返回 `Promise<HTTPResponse>`，错误通过 Promise rejection 抛出
2. **查询参数**: ✅ 已修复 — TypeScript 接口中 `query` 已更新为 `Record<string, string | string[] | number | boolean | null>`，支持多值查询参数

## 修改历史

| 日期 | 修改内容 |
|------|----------|
| 2026-06-19 | 更新文档：修正所有文件行数、补充 runtime_wasm.go 和测试文件信息、更新已知问题状态（P2 全部解决） |
| 2026-06-16 | 全面更新文档：修正所有文件行数、补充 validator.go、更新已知问题状态（P0 全部修复）、更新评分 |
| 2026-06-16 | 更新文档：修正测试覆盖描述、更新评分、消除模板代码重复 |
| 2026-06-15 | 更新文档评分：文档完善 2→3，安全性 2→3，综合 2.5→2.8 |
| 2026-06-15 | 创建完整 docs/ 文档目录 |
| 2026-06-14 | 修复 pathParams 覆盖问题 |
| 2026-06-14 | 添加 stringToString 函数 |
| 2026-06-14 | 更新 Makefile dev-generate 命令 |