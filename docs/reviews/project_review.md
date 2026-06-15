基于对本项目源代码的深入探测、架构梳理以及与 `README.md` 规范的对比，生成的多维度深度评审报告如下：

本项目整体设计思路清晰，巧妙地利用 WASM 为前端提供 Go 语言级别的 HTTP 客户端和 API SDK 代码生成。但在实际落地的细节中，存在一些**严重的功能脱节、并发安全问题以及实现上的 Bug**。

以下是具体的评审报告：

### 1. 一致性核对（实际代码 vs README.md）

*   **❌ WASM 自动编译机制失效**
    *   **描述**：`README.md` 承诺“Auto-compile WASM after code generation”。但实际在 `cmd/generator/main.go` 中，编译指令将生成的目标目录（`outDir`）作为 package path 传递给 `go build`。由于生成目录只包含 `generated.go`，且既没有 `main.go` 入口文件也没有 `go.mod`（未调用生成），这必然会导致 `package is not in std` 的编译报错。
*   **❌ Go 侧类型安全验证成为“死代码”**
    *   **描述**：`README.md` 声称提供端到端的 Type-safe SDK 以及 Request Validation。生成的 `generated.go` 确实包含了完美的结构体和 `init()` 路由注册逻辑。但是，WASM 运行时的 JS 接口点 (`exports.go` 中的 `callAPI`) 直接**忽略了传入的 `operationId`**（代码中有明确注释 `_ = args[0].String() // operationId for future use`），然后直接使用了底层通用的 `client.Call` 兜底发送。这导致生成的 Go 结构体、校验逻辑（`Validate()`）根本没有被集成和调用。
*   **❌ Makefile 构建脚本过时**
    *   **描述**：CLI 工具在重构时启用了子命令结构（如 `gowasm-generator generate -s ...`），但 `Makefile` 中对应的 `make dev-generate` 仍在执行旧的无子命令格式（`go run ./cmd/generator -spec=...`），导致开箱运行报错。

### 2. 代码质量与架构分析

*   **架构割裂问题 (Architecture Disconnect)**：
    *   当前系统分为 `runtime` 与 `generator` 两部分。生成器输出了 `generated.go` 代码，但官方默认编译为 WASM 的入口是 `cmd/runtime/main.go`，该入口仅导入了泛用的底层客户端（`ExportMain`），**完全没有 import 用户生成的代码**。两端高度解耦导致生成代码其实在浏览器内根本不生效。
*   **漏调的 `go.mod` 生成逻辑**：
    *   在 `pkg/generator/go_templates.go` 中定义了完整的 `writeGoMod` 方法和 `go.mod.tmpl` 模板，但在 `generator.go` 的主入口 `Generate()` 中遗漏了对它的调用，这阻碍了生成的 Go 代码作为一个独立模块被编译。
*   **查询参数 (Query/Path) 模型设计缺陷**：
    *   在 Go 的 `Request` 结构体和生成的 TypeScript 接口中，`query` 字段都被强约束为 `map[string]string`。这意味着**不支持同名多值的数组查询参数**（例如 `?status=available&status=pending`），在实际业务场景中（如 Petstore 的 `findPetsByStatus`）会遇到解析瓶颈。强制将所有 Request Body 转义为 `application/json` 也不支持诸如 `FormData` 类型的提交。
*   **Go 模板编译时潜在语法错误**：
    *   在 `pkg/generator/templates/sdk.go.tmpl` 处理路径参数映射时，模板写死了 `{{.GoType}}ToString(params.{{.GoName}})`。如果该字段原本就是 `string` 类型，就会生成调用 `stringToString()`。由于模板末尾只定义了 `intToString` 等助手方法而遗漏了 `string`，一旦有字符串参数即可导致用户侧编译 Go 代码失败。

### 3. 潜在风险与漏洞

*   **🚨 严重的并发数据竞争引发 Panic (Data Race)**：
    *   在 `pkg/runtime/client.go` 中，`HTTPClient` 的配置项 `config.Headers` 结构是一个原生的 `map[string]string`。
    *   当在 JS 侧调用 `wasmCallAPI` 时，`exports.go` 会**新开一个 goroutine** 并发执行 HTTP 请求（并在请求中遍历读取 `Headers`）。
    *   此时若 JS 主线程调用了 `wasmSetAuthToken()`，该方法会同步且无锁地执行 `client.SetAuthToken` 写入 Map。Go 的 Map 非并发安全，**极易引发原生的 Concurrent map iteration and map write 级 Panic**，直接导致整个 WASM 实例彻底崩溃。
*   **🚨 路径参数插值完全瘫痪 (Path Parameters Ignored)**：
    *   这是一个阻断性 Bug。在 `exports.go` 读取 JS 对象 (`reqJS`) 的流程中，完全没有解析传入的 `pathParams`。
    *   更糟糕的是，在底层 `client.go` 的 `buildURL` 方法中，拼接路径的代码为 `fullURL := ResolvePath(path, nil, query)`，**硬编码传了 `nil`** 给路径参数替换逻辑。
    *   这意味着所有带有路径参数的请求（如 `/pet/{petId}`）根本不会被替换，而是直接把包含花括号的原始字符串打到了远端服务器。
*   **静默失败的路径防御**：
    *   在 `ResolvePath` 的 `safePathParam` 防御路径穿越漏洞（如 `../`）时，检测到非法输入会直接返回 `""`。这种将参数静默抹除的策略，可能会让前端难以察觉，最终拼凑出如 `/users//profile` 等错误的 URL，更推荐直接抛出明显的错误异常。

### 4. 具体改进建议

基于以上评审结果，建议进行如下高优先级的调整和重构：

**第一阶段：修复阻断性 Bug 与高危风险（紧急）**
1.  **修复并发崩溃风险**：在 `ClientConfig` 或 `HTTPClient` 内引入 `sync.RWMutex`，对 `config.Headers` 的所有读（请求构造）和写（`SetAuthToken`、`ClearAuthToken`）进行加锁保护。
2.  **打通路径参数 (PathParams)**：
    *   在 `exports.go` 解析 `reqJS` 时补充解析 `pathParams` 结构为 `map[string]string`。
    *   修改 `client.Call` 让其把 `req.PathParams` 传递给 `buildURL`，将替换功能激活。
3.  **修补 Go 模板**：在 `sdk.go.tmpl` 内加入 `func stringToString(v string) string { return v }`，防止含有 String 路径参数的 API 生成出无法编译的代码。

**第二阶段：修复架构与功能一致性（重要）**
4.  **修正 WASM 的打包与架构关系**：
    *   **路线A (通用客户端)**：若定位是只要通用 HTTP 调用，请将 `README.md` 中关于 Go Typesafe 特性说明移除，并移除 `generated.go` 的生成逻辑，仅保留 `sdk.ts` 即可。
    *   **路线B (强类型拦截)**：若要实现完整的操作路由验证，需要让生成器输出一份 `main.go`（里面引入 `generated.go` 和 `runtime`），然后再将该 `main.go` 所在目录传递给 `go build` 构建。修改 `exports.go`，让 `callAPI` 通过 `operationID` 去 `GetOperation(opId)` 并执行相应的回调。
5.  **补齐缺失的执行逻辑**：
    *   在 `Generator.Generate` 执行流中加上对 `writeGoMod` 的调用。
    *   修改 `Makefile` 中的 `dev-generate` 命令，加入 `generate -s ...` 以兼容新的 CLI 架构。

**第三阶段：扩展与优化（提升）**
6.  **升级请求参数设计**：将 HTTP 客户端底层的 `Query map[string]string` 重构为 `url.Values`，以良好地支持多个相同的 Query 参数名。
7.  **增强 Request Body 的拓展性**：移除仅支持 JSON 序列化的强制处理，允许透传 `ArrayBuffer` 或者 `Blob` 等，使其在应对各种类型的网络提交（文件上传，Form提交等）具有更好的灵活性。