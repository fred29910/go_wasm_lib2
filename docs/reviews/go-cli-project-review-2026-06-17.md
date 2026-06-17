# Go CLI 项目评审报告

评审对象：`go_wasm_lib2`

评审日期：2026-06-17

评审范围：

- CLI 规范与交互设计
- 功能与逻辑实现
- Go 代码质量与目录结构
- 错误处理与日志记录
- 依赖与性能优化
- 生成 SDK 质量
- 测试覆盖与工程化成熟度

本报告记录一次面向 Go CLI / WASM SDK Generator 项目的详细 Code Review 与项目评审结论。结论基于当前工作区源码、文档、CLI 行为、生成样例和基础验证命令。

---

## 1. 总体结论

该项目是一个 Go/WASM HTTP SDK 生成器，主要能力包括：

- 从 OpenAPI 3.x YAML 解析 API 模型；
- 生成 Go client 代码；
- 生成 TypeScript SDK；
- 生成调试用 HTML；
- 生成 WASM runtime 入口；
- 支持标准 Go 与 TinyGo 构建 WASM；
- 支持生成后运行 oxlint。

项目已经具备较好的雏形，并且已经覆盖了不少关键安全问题，例如：

- path parameter 确定性替换；
- path traversal 检测；
- response body 10MB 读取上限；
- `WASMError` 结构化错误；
- 生成 `APIClient` 替代早期全局单例；
- 生成器与 runtime 的基础测试。

但当前项目距离“生产级 CLI / SDK generator”仍有明显差距。最需要优先修复的问题集中在：

1. 生成产物的可用性存在现成问题；
2. CLI 架构偏单文件大 Action；
3. `pkg/runtime` 职责过重；
4. required 字段校验和 TS 类型存在明显逻辑缺口；
5. CLI 文档、Taskfile、实际命令之间存在不一致；
6. 生成 SDK 的 `go.mod`、`cmd/wasm/main.go`、required validation、WASM loading 需要立即修正。

本次评审没有修改业务代码，只新增本评审文档到 `docs/reviews/`。

---

## 2. 验证记录

本次评审执行了以下验证命令：

```bash
go test ./...
```

结果：通过。

```bash
go test -race ./...
```

结果：通过。

```bash
go vet ./cmd/... ./pkg/... ./version/...
```

结果：通过。

```bash
GOOS=js GOARCH=wasm go build -trimpath -o /tmp/gowasm-review-runtime.wasm ./cmd/runtime
```

结果：通过。

CLI help 验证：

```bash
go run ./cmd/generator --help
go run ./cmd/generator generate --help
go run ./cmd/generator init --help
```

结果：help 可正常输出，但缺少 `Examples`、`ArgsUsage`、`version` 命令等体验增强项。

生成样例验证：

```bash
go run ./cmd/generator generate \
  -s examples/petstore/openapi.yaml \
  -o /tmp/gowasm-review-generated \
  --wasm=false \
  --oxlint-disable
```

结果：生成成功，但生成的 `go.mod` 模块路径错误：

```go
module /tmp/gowasm-review-generated
```

这会导致：

```bash
go test ./...
```

失败：

```text
go: malformed module path "/tmp/gowasm-review-generated": empty path element
```

进一步验证：

```bash
go run ./cmd/generator generate \
  -s examples/petstore/openapi.yaml \
  -o /tmp/gowasm-review-generated3 \
  --module github.com/acme/sdk \
  --wasm=false \
  --oxlint-disable
```

结果：虽然传入了 `--module github.com/acme/sdk`，生成的 `go.mod` 仍然是：

```go
module /tmp/gowasm-review-generated3
```

说明 `--module` 当前没有真正作用于生成模块路径。

手动修正模块路径并执行：

```bash
go mod tidy
GOOS=js GOARCH=wasm go test ./...
```

结果：WASM target 下生成代码可通过编译测试。

但原生平台执行：

```bash
go test ./...
```

仍失败，原因是生成的 `cmd/wasm/main.go` 缺少：

```go
//go:build js && wasm
```

并且该文件引用了只在 WASM build 下存在的 `runtime.ExportMain()`。

---

## 3. 项目结构概览

当前主要结构：

```text
cmd/
  generator/main.go
  runtime/main.go

pkg/
  generator/
    generator.go
    openapi.go
    types.go
    go_templates.go
    ts_templates.go
    templates/
  runtime/
    client.go
    error.go
    exports.go
    promise.go
    converter.go
    validator.go
    build.go

version/
  version.go

docs/
  cli-reference.md
  generator-api.md
  runtime-api.md
  architecture.md
  getting-started.md
  known-issues.md
```

当前 CLI 框架：

```go
github.com/urfave/cli/v2 v2.27.7
```

当前没有使用 Cobra 或 Viper。`urfave/cli/v2` 对当前规模是合理选择，但如果未来命令树继续扩展，需要重新评估 CLI 架构。

---

## 4. 问题分级总览

| 级别 | 问题 | 影响 | 建议优先级 |
|---|---|---|---|
| P0 | `--module` 不生效，生成 `go.mod` 使用输出目录作为 module path | 生成 SDK 无法作为正常 Go module 使用 | 立即修复 |
| P0 | 生成的 `cmd/wasm/main.go` 缺少 `//go:build js && wasm` | 原生平台 `go test ./...` 失败 | 立即修复 |
| P0 | 生成的 `cmd/wasm/main.go` 硬编码 runtime import | 自定义 `RuntimeImport` 不生效 | 立即修复 |
| P0 | required 字段使用零值判断 | `0`、`false` 等合法值会被误判为缺失 | 立即修复 |
| P0 | `GetConfig()` 返回内部 config 指针 | 可能导致 data race 和外部误改 | 立即修复 |
| P0 | path param 非法时静默变成空字符串 | 错误不清晰，调试困难 | 立即修复 |
| P0 | JSON 模式直接 `os.Exit(1)` | 绕过 CLI 框架，难以测试 | 立即修复 |
| P1 | `pkg/runtime` 是 god package | 职责混杂，测试和维护困难 | 下一轮重构 |
| P1 | `cmd/generator/main.go` 过厚 | CLI 逻辑难以拆分和测试 | 下一轮重构 |
| P1 | TS `query` 类型与 runtime `url.Values` 不一致 | 多值查询无法正确表达 | 下一轮修复 |
| P1 | TS `WASMSDK.load()` 与 Go WASM runtime 加载方式不匹配 | 浏览器加载可能失败 | 下一轮修复 |
| P1 | build/tidy 无 timeout/context | 外部命令卡住时 CLI 可能挂起 | 下一轮修复 |
| P2 | 缺少结构化日志 | 生产排障能力不足 | 中期优化 |
| P2 | OpenAPI 支持范围较窄但文档不够明确 | 用户预期不稳定 | 中期优化 |
| P2 | CLI help 缺少 examples/args usage | 用户体验一般 | 中期优化 |
| P2 | 文档与 CLI 不一致 | 降低文档可信度 | 中期优化 |

---

## 5. CLI 规范与交互设计评审

### 5.1 CLI 框架选择

当前使用：

```go
github.com/urfave/cli/v2
```

位置：`go.mod:5-9`

评价：

- 对当前项目规模足够；
- 比 Cobra 更轻量；
- 不引入 Viper 是合理的，因为当前没有配置文件、多 profile、环境变量合并等复杂需求。

建议：

- 如果继续走轻量 CLI 路线，应把 `urfave/cli/v2` 用规范；
- 如果未来命令树扩展到多级，例如 `sdk generate`、`sdk build`、`sdk init`、`sdk doctor`，再考虑迁移到 Cobra；
- 不建议为了“标准”强行引入 Viper。

### 5.2 命令结构

当前命令：

```go
gowasm-generator generate
gowasm-generator init
```

位置：`cmd/generator/main.go:36-128`

优点：

- 命令数量少，学习成本低；
- `generate` 与 `init` 职责基本清晰；
- `--help` 可用。

问题：

- 缺少 `Examples`；
- 缺少 `ArgsUsage`；
- 没有配置 shell completion；
- 文档中写了 `version` 命令，但实际 CLI 没有该命令。

文档位置：`docs/cli-reference.md:13`

建议：

```go
{
    Name:      "generate",
    Usage:     "Generate Go and TypeScript SDK from an OpenAPI spec",
    ArgsUsage: "--spec <file> [--out <dir>] [flags]",
    Examples: `gowasm-generator generate -s api/openapi.yaml -o ./generated --wasm=false
gowasm-generator generate -s api/openapi.yaml -o ./generated --dry-run --output json`,
    Action: runGenerate,
}
```

### 5.3 `-v` alias 冲突

当前：

- root command 自动提供 `--version, -v`；
- generate command 提供 `--validation, -v`。

位置：`cmd/generator/main.go:94-99`

问题：

```bash
gowasm-generator -v
gowasm-generator generate -v
```

两个 `-v` 含义不同，容易造成用户困惑。

建议：

- 移除 `--validation` 的 `-v` alias；
- 或改成 `--no-validation`，默认 true。

### 5.4 `--wasm` 默认 true 风险较高

当前：

```go
&cli.BoolFlag{
    Name:  "wasm",
    Usage: "Build WASM after generation",
    Value: true,
}
```

位置：`cmd/generator/main.go:80-84`

问题：

- 用户只想生成代码时，会默认触发 `go mod tidy`、WASM build、oxlint；
- CI 中可能因为 TinyGo、Node、npx 或网络依赖失败；
- `--dry-run --wasm=true` 语义不自然。

建议二选一：

#### 方案 A：默认不构建 WASM

```go
&cli.BoolFlag{
    Name:  "wasm",
    Usage: "Build WASM after generation",
    Value: false,
}
```

#### 方案 B：拆成独立命令

```bash
gowasm-generator generate ...
gowasm-generator build-wasm --out ./generated
```

更推荐方案 B，尤其是未来构建参数继续增加时。

### 5.5 JSON 模式直接 `os.Exit(1)`

当前多处：

```go
if outputFormat == "json" {
    outputJSON(JSONOutput{
        Success: false,
        Error:   err.Error(),
    })
    os.Exit(1)
}
return err
```

位置：

- `cmd/generator/main.go:188-193`
- `cmd/generator/main.go:220-225`
- `cmd/generator/main.go:239-244`
- `cmd/generator/main.go:255-260`

问题：

- `os.Exit` 会绕过 deferred cleanup；
- CLI 框架无法统一处理 exit code；
- 单元测试难以覆盖；
- JSON 输出后是否还会打印框架错误，取决于框架行为，容易不稳定。

建议：

- Action 只 return error；
- 使用 urfave/cli 的 `ExitErrHandler` 或 `cli.Exit` 统一处理；
- JSON failure 写成 helper。

---

## 6. 功能与逻辑实现评审

### 6.1 `--module` 不生效，生成 `go.mod` 使用输出目录

当前配置：

```go
type Config struct {
    ModuleName    string
    OutputModule  string
    Package       string
    RuntimePath   string
    RuntimeImport string
}
```

位置：`pkg/generator/generator.go:14-24`

生成模板：

```go
module {{.Config.OutputModule}}
```

位置：`pkg/generator/templates/go.mod.tmpl:1`

但 CLI 中：

```go
if outDir != "" {
    cfg.OutputModule = outDir
}
```

位置：`cmd/generator/main.go:157-162`

结果：

```bash
go run ./cmd/generator generate \
  -s examples/petstore/openapi.yaml \
  -o /tmp/gowasm-review-generated \
  --module github.com/acme/sdk \
  --wasm=false \
  --oxlint-disable
```

生成的 `go.mod` 是：

```go
module /tmp/gowasm-review-generated
```

而不是：

```go
module github.com/acme/sdk
```

建议重构：

#### 修改前

```go
type Config struct {
    ModuleName    string
    OutputModule  string
    Package       string
}
```

```go
if outDir != "" {
    cfg.OutputModule = outDir
}
```

```go
module {{.Config.OutputModule}}
```

#### 修改后

```go
type Config struct {
    ModuleName string
    OutputDir  string
    Package    string
}
```

```go
if outDir != "" {
    cfg.OutputDir = outDir
}
```

```go
module {{.Config.ModuleName}}
```

并且 `Generate` 使用 `cfg.OutputDir` 写文件。

### 6.2 生成的 `cmd/wasm/main.go` 缺少 WASM build tag

当前模板：

```go
package main

import (
    "github.com/fred29910/gowasm/pkg/runtime"
)

func main() {
    runtime.ExportMain()
}
```

位置：`pkg/generator/templates/main.go.tmpl:3-10`

问题：

1. 缺少：

```go
//go:build js && wasm
```

2. 硬编码：

```go
"github.com/fred29910/gowasm/pkg/runtime"
```

没有使用 `{{.Config.RuntimeImport}}`。

建议：

```go
// Code generated by github.com/fred29910/gowasm/cmd/generator. DO NOT EDIT.

//go:build js && wasm

package main

import runtime "{{.Config.RuntimeImport}}"

func main() {
    runtime.ExportMain()
}
```

### 6.3 `GenerateDryRun` 与 `Generate` 文件列表不一致

实际 dry-run 输出：

```text
Dry run: files that would be generated:
  generated.go (6270 bytes)
  sdk.ts (4731 bytes)
  index.html (30139 bytes)

Total: 3 file(s), 41140 bytes
```

但实际 `Generate` 会写：

- `generated.go`
- `go.mod`
- `cmd/wasm/main.go`
- `sdk.ts`
- `index.html`

位置：

- `pkg/generator/generator.go:168-222`
- `docs/cli-reference.md:142-150`

建议：

`GenerateDryRun` 应渲染并统计所有会生成的文件，包括：

- `generated.go`
- `go.mod`
- `cmd/wasm/main.go`
- `sdk.ts`
- `index.html`

### 6.4 required 字段校验使用零值判断

当前模板：

```go
if r.PetID == int64(0) {
    return fmt.Errorf("petId is required")
}
```

位置：`pkg/generator/templates/sdk.go.tmpl:76-79`

问题：

- `0` 可能是合法值；
- `false` 可能是合法值；
- required body 没有校验；
- Petstore 的 `createPet` requestBody 是 required，但生成的 `CreatePetRequest.Validate()` 返回 nil。

建议 required 字段使用 pointer。

#### 修改前

```go
type GetPetByIDRequest struct {
    PetID int64 `json:"petId"`
}

func (r GetPetByIDRequest) Validate() error {
    if r.PetID == int64(0) {
        return fmt.Errorf("petId is required")
    }
    return nil
}
```

#### 修改后

```go
type GetPetByIDRequest struct {
    PetID *int64 `json:"petId"`
}

func (r GetPetByIDRequest) Validate() error {
    if r.PetID == nil {
        return fmt.Errorf("petId is required")
    }
    return nil
}
```

对 body required 也应类似：

```go
Body *Pet `json:"body,omitempty"`
```

### 6.5 required 字段仍带 `omitempty`

当前模板：

```go
{{.GoName}} {{.GoType}} `json:"{{.Name}},omitempty"`
```

位置：`pkg/generator/templates/sdk.go.tmpl:18`

问题：

- required 字段不应该默认 `omitempty`；
- required 字段如果为零值，JSON marshal 会省略，但业务上它应该是 required；
- 如果不用 pointer，这个问题无法彻底解决。

建议：

- required 字段使用 pointer；
- required 字段不要 `omitempty`；
- optional 字段才 `omitempty`。

### 6.6 path param 非法时静默变成空字符串

当前：

```go
func safePathParam(v string) string {
    unescaped, err := url.PathUnescape(v)
    if err != nil {
        return ""
    }
    if strings.Contains(unescaped, "..") ||
        strings.Contains(unescaped, "//") ||
        strings.HasPrefix(unescaped, "/") {
        return ""
    }
    return v
}
```

位置：`pkg/runtime/client.go:248-257`

问题：

- 非法 path param 被替换为空字符串；
- 用户收到的是 malformed URL，而不是明确错误；
- 测试甚至把这个行为固定为期望值。

建议：

`ResolvePath` 应返回 error。

#### 修改前

```go
func ResolvePath(path string, pathParams map[string]string, query url.Values) string
```

#### 修改后

```go
var ErrInvalidPathParam = errors.New("invalid path parameter")

func ResolvePath(path string, pathParams map[string]string, query url.Values) (string, error) {
    resolved, err := resolvePathParams(path, pathParams)
    if err != nil {
        return "", err
    }

    if query == nil || len(query) == 0 {
        return resolved, nil
    }

    sep := "?"
    if strings.Contains(resolved, "?") {
        sep = "&"
    }

    return resolved + sep + query.Encode(), nil
}
```

### 6.7 `GetConfig()` 返回内部 config 指针

当前：

```go
func (c *HTTPClient) GetConfig() *ClientConfig {
    return c.config
}
```

位置：`pkg/runtime/client.go:292-294`

问题：

- 调用方可以修改 `Headers`、`Timeout`、`BaseURL`；
- 与 `SetAuthToken` 并发时可能产生 data race；
- `ClientConfig.Headers` 是 map，返回引用尤其危险。

建议：

```go
func (c *HTTPClient) GetConfig() *ClientConfig {
    c.mu.RLock()
    defer c.mu.RUnlock()

    headers := make(map[string]string, len(c.config.Headers))
    for k, v := range c.config.Headers {
        headers[k] = v
    }

    return &ClientConfig{
        BaseURL:     c.config.BaseURL,
        Timeout:     c.config.Timeout,
        Headers:     headers,
        Credentials: c.config.Credentials,
    }
}
```

### 6.8 `NewHTTPClient` 对 nil Headers 没有保护

当前：

```go
return &HTTPClient{
    config: config,
    httpClient: &http.Client{
        Timeout: timeout,
    },
}
```

位置：`pkg/runtime/client.go:80-96`

如果用户传入：

```go
&runtime.ClientConfig{BaseURL: "..."}
```

`Headers` 是 nil。

之后调用：

```go
client.SetAuthToken("token", "Bearer")
```

会 panic。

建议：

```go
if config.Headers == nil {
    config.Headers = make(map[string]string)
}
```

### 6.9 `context.Background()` 用于 WASM API 调用

当前：

```go
ctx := context.Background()
```

位置：`pkg/runtime/exports.go:186`

问题：

- JS 侧 abort 后，Go/WASM 侧请求无法取消；
- 页面卸载或 SDK 重新 init 时，in-flight request 可能继续跑；
- timeout 只来自 `http.Client.Timeout`，不是 request context。

建议：

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Duration(client.GetConfig().Timeout)*time.Second)
defer cancel()
```

更完整方案：JS 侧支持 AbortController，并映射到 Go context。

### 6.10 `BuildWASM` / `RunModTidy` 没有超时

当前：

```go
cmd := exec.Command("tinygo", "build", "-o", outPath, "-target=wasm", ".")
```

位置：`pkg/runtime/build.go:108-120`

以及：

```go
cmd := exec.Command("go", "mod", "tidy")
```

位置：`pkg/runtime/build.go:143-149`

问题：

- 编译器或 `go mod tidy` 卡住时，CLI 可能无限挂起；
- 没有 context；
- 没有超时策略。

建议：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

cmd := exec.CommandContext(ctx, "tinygo", "build", "-o", outPath, "-target=wasm", ".")
```

---

## 7. Go 代码质量与目录结构评审

### 7.1 `pkg/runtime` 是 god package

当前 `pkg/runtime` 同时包含：

- HTTP client：`client.go`
- WASM JS exports：`exports.go`
- JS/Go converter：`converter.go`
- Promise helper：`promise.go`
- validators：`validator.go`
- WASM error：`error.go`
- build orchestration：`build.go`

问题：

- 包职责过重；
- 测试边界不清晰；
- WASM-only 代码与非 WASM 代码混在一起；
- 生成 SDK 只需要 runtime client，却被迫依赖整个 runtime 包。

建议拆分：

```text
pkg/runtime/
  client/        # HTTPClient, Request, Response, ClientConfig
  errors/        # WASMError, WrapWASMError
  validate/      # IsValidEmail, IsValidUUID, IsValidDateTime
  wasm/          # JS exports, PromiseHelper
  convert/       # JS/Go converter
  build/         # BuildWASM, DetectTinyGo
```

如果担心生成 SDK import 路径变化，可以保留 `pkg/runtime` 作为 facade：

```go
package runtime

import "github.com/fred29910/gowasm/pkg/runtime/client"

type HTTPClient = client.HTTPClient
type Request = client.Request
type Response = client.Response
```

### 7.2 `cmd/generator/main.go` 过厚

当前 `runGenerate` 同时负责：

- 读取 flags；
- 构造 generator config；
- 执行 generation；
- 输出 text/json；
- 构建 WASM；
- 运行 oxlint；
- 处理错误和退出码。

位置：`cmd/generator/main.go:135-284`

建议拆成：

```text
cmd/generator/
  main.go
  generate.go
  init.go
  flags.go
  output.go
  wasm.go
  lint.go
```

或者：

```text
internal/cli/generate.go
internal/cli/init.go
internal/cli/output.go
```

`cmd/generator/main.go` 应该只负责 app 装配。

### 7.3 `pkg/generator.Config` 字段语义混乱

当前：

```go
ModuleName    string // Go module name for generated code
OutputModule  string // Output module path
Package       string // Go package name
```

位置：`pkg/generator/generator.go:14-20`

但实际：

- `--module` 设置 `cfg.ModuleName`；
- `--out` 设置 `cfg.OutputModule`；
- 模板使用 `OutputModule` 作为 `go.mod` 的 module path。

建议：

```go
type Config struct {
    ModuleName     string
    OutputDir      string
    Package        string
    RuntimePath    string
    RuntimeImport  string
    Validation     bool
    GoTemplatePath string
    TSTemplatePath string
    Verbose        bool
}
```

### 7.4 OpenAPI 支持范围较窄

当前 parser 支持：

- paths；
- components.schemas；
- basic schema type/format/enum；
- internal `$ref`；
- requestBody；
- responses。

不支持或支持不完整：

- `allOf`；
- `oneOf`；
- `anyOf`；
- discriminator；
- external `$ref`；
- path-level parameters；
- callbacks；
- security schemes；
- JSON Schema constraints；
- nullable 的 pointer 语义；
- request body 多 media type；
- response 多 media type。

位置：`pkg/generator/openapi.go:13-80`

建议：

- 明确文档声明“当前支持 OpenAPI 3.x 子集”；
- 对不支持特性返回清晰错误，而不是静默降级为 `interface{}` / `any`。

### 7.5 response 主响应选择依赖 map 遍历

当前：

```go
for code := range op.Responses {
    if code >= "200" && code < "300" {
        primaryCode = code
        break
    }
}
```

位置：`pkg/generator/generator.go:343-357`

问题：

- map 遍历顺序随机；
- 如果多个 2xx response，主响应不 deterministic；
- 字符串比较状态码不够严谨。

建议：

- 收集所有 response codes；
- 排序；
- 优先选择最低 2xx；
- 否则选择最低可用 status code。

---

## 8. 错误处理与日志记录评审

### 8.1 `WASMError` 设计方向正确

`WASMError` 支持：

- code；
- message；
- details；
- filePath；
- lineNumber；
- suggestion；
- Unwrap。

位置：`pkg/runtime/error.go:10-47`

建议保留该设计，并继续完善 JS 侧暴露字段。

### 8.2 Promise rejection 丢失 suggestion

当前：

```go
jsErr.Set("code", wasmErr.Code)
if wasmErr.Details != "" {
    jsErr.Set("details", wasmErr.Details)
}
```

位置：`pkg/runtime/promise.go:45-55`

没有设置：

- `suggestion`；
- `filePath`；
- `lineNumber`。

建议：

```go
jsErr.Set("code", wasmErr.Code)
jsErr.Set("details", wasmErr.Details)
jsErr.Set("suggestion", wasmErr.Suggestion)
jsErr.Set("filePath", wasmErr.FilePath)
jsErr.Set("lineNumber", wasmErr.LineNumber)
```

### 8.3 Promise executor 没有 recover

当前：

```go
executor(resolve, reject)
```

位置：`pkg/runtime/promise.go:20-30`

如果 executor 中 panic，goroutine 会崩溃，Promise 不会 resolve/reject。

建议：

```go
func (p *PromiseHelper) CreatePromise(executor func(resolve, reject js.Value)) js.Value {
    promiseConstructor := js.Global().Get("Promise")
    handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        if len(args) != 2 {
            return nil
        }

        resolve := args[0]
        reject := args[1]

        go func() {
            defer func() {
                if r := recover(); r != nil {
                    p.RejectPromise(reject, fmt.Errorf("panic in WASM promise executor: %v", r))
                }
            }()

            executor(resolve, reject)
        }()

        return nil
    })

    return promiseConstructor.New(handler)
}
```

### 8.4 `CreateRejectedPromise` 对 nil error 不安全

当前：

```go
jsErr := js.Global().Get("Error").New(err.Error())
```

位置：`pkg/runtime/promise.go:79-82`

如果 `err == nil`，会 panic。

建议：

```go
func (p *PromiseHelper) CreateRejectedPromise(err error) js.Value {
    promiseConstructor := js.Global().Get("Promise")
    if err == nil {
        err = errors.New("unknown error")
    }
    jsErr := js.Global().Get("Error").New(err.Error())
    return promiseConstructor.Call("reject", jsErr)
}
```

### 8.5 CLI 错误输出基本合规，但 JSON 模式有隐患

当前顶层：

```go
if err := app.Run(os.Args); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

位置：`cmd/generator/main.go:129-132`

优点：

- 普通错误输出到 stderr；
- 退出码为 1。

问题：

- JSON 模式中直接 `os.Exit(1)`，绕过统一错误处理；
- 没有 `ExitErrHandler`；
- 没有区分 usage error 和 runtime error。

建议：

- 配置 `app.ExitErrHandler`；
- usage error 返回 usage help；
- runtime error 不重复打印 usage；
- JSON 模式只输出 JSON，不输出额外文本。

### 8.6 没有结构化日志

当前主要是：

- `fmt.Printf`；
- `fmt.Println`；
- `fmt.Fprintln(os.Stderr, ...)`。

位置：

- `pkg/runtime/build.go:109`
- `pkg/runtime/build.go:125`
- `pkg/generator/generator.go:75-79`
- `cmd/generator/main.go` 多处

对于当前规模可以接受，但如果 CLI 要进入生产使用，建议引入 Go 标准库：

```go
log/slog
```

---

## 9. 依赖与性能优化评审

### 9.1 依赖选择总体合理

当前依赖：

```go
github.com/urfave/cli/v2
gopkg.in/yaml.v3
github.com/norunners/vert
```

位置：`go.mod:5-9`

| 依赖 | 评价 |
|---|---|
| `urfave/cli/v2` | 合理，轻量稳定 |
| `gopkg.in/yaml.v3` | 合理，Go YAML 常见选择 |
| `github.com/norunners/vert` | 可用但维护活跃度需关注 |

`norunners/vert` 用于 JS/Go 值转换，是核心路径，但维护活跃度不如主流 JSON 库。建议评估是否可以逐步替换为手写转换加 `encoding/json`。

### 9.2 `npx oxlint` 默认运行有外部副作用

当前：

```go
cmd := exec.Command("npx", "oxlint", "-c", configFile, "--no-ignore", targetDir)
```

位置：`cmd/generator/main.go:336-347`

问题：

- `npx` 可能触发网络安装；
- 用户环境没有 Node.js 时会失败；
- 生成 SDK 的核心流程依赖前端 lint 工具。

建议：

- 默认不 lint；
- 或提供 `--lint-ts` 显式开启；
- 或把 oxlint 作为独立命令：

```bash
gowasm-generator lint-ts --dir ./generated
```

### 9.3 `go mod tidy` 在生成流程中自动执行

当前：

```go
if err := runtime.RunModTidy(outDir); err != nil {
    return fmt.Errorf("go mod tidy failed: %w", err)
}
```

位置：`cmd/generator/main.go:310-314`

问题：

- 生成流程隐式修改 `go.mod` / `go.sum`；
- 用户可能只想 dry-run 或只生成代码；
- `go mod tidy` 可能比较慢。

建议：

- 增加 `--mod-tidy` 或 `--no-mod-tidy`；
- 如果 `--wasm=false`，是否 tidy 应明确文档化。

---

## 10. 生成 SDK 专项评审

### 10.1 TypeScript `query` 类型与 runtime 不一致

runtime 支持多值查询：

```go
Query url.Values
```

位置：`pkg/runtime/client.go:99-107`

但生成的 TS：

```ts
query?: Record<string, string>;
```

位置：`pkg/generator/templates/sdk.ts.tmpl:20-22`

建议：

```ts
query?: Record<string, string | string[] | number | boolean | null>;
```

或者更严格：

```ts
type QueryValue = string | string[];
query?: Record<string, QueryValue>;
```

同时 TS 生成逻辑需要把 array 展开。

### 10.2 TypeScript 字段命名存在不自然结果

实际生成：

```ts
export interface Pet {
    iD?: number;
    name: string;
    status?: 'available' | 'pending' | 'sold';
}
```

`id` 变成 `iD`，`petId` 变成 `petID`。

这不是致命错误，但对 TS 用户不友好。

建议调整 `ToTSName`：

- 对 acronym 做整体 lowercase 或保留用户原始 casing；
- 或提供更可预测规则：

```text
id -> id
userId -> userId
petId -> petId
URLPath -> urlPath
```

### 10.3 TS SDK 的 `load()` 与 Go WASM 实际加载方式不匹配

当前：

```ts
const wasmModule = await WebAssembly.instantiate(bytes, {
    env: {
        // Go WASM imports
    }
});
```

位置：`pkg/generator/templates/sdk.ts.tmpl:89-109`

Go WASM 通常需要 `wasm_exec.js` 提供的 `Go` runtime：

```js
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch(this.wasmUrl),
    go.importObject,
);
go.run(result.instance);
```

当前 `index.html` 里反而有更接近正确的实现。

建议：

- 统一 TS SDK 和 `index.html` 的加载逻辑；
- 明确 `wasm_exec.js` 必须和 `main.wasm` 一起部署；
- 如果 `load()` 不负责加载 `wasm_exec.js`，文档必须说明。

### 10.4 生成的 Go client 和 runtime operation registry 没有真正打通

runtime 支持：

```go
RegisterOperation(operationID, handler)
GetOperation(operationID)
```

位置：`pkg/runtime/client.go:49-65`

但生成的 `APIClient` 只是自己维护 operations map：

```go
operations map[string]runtime.OperationHandler
```

位置：`pkg/generator/templates/sdk.go.tmpl:37-43`

生成方法直接调用：

```go
return c.client.Call(ctx, &req)
```

位置：`pkg/generator/templates/sdk.go.tmpl:123-127`

这意味着：

- JS 侧 `wasmCallAPI(operationId, request)` 不会自动路由到生成的 typed method；
- `APIClient.operations` 基本是 dead code；
- 如果希望 operationId 路由生效，生成器需要生成 `runtime.RegisterOperation(...)`。

建议至少二选一：

#### 方案 A：移除 `APIClient.operations`

如果当前设计就是 typed Go client 独立调用，不要保留误导性 registry。

#### 方案 B：生成 `runtime.RegisterOperation`

例如：

```go
func init() {
    runtime.RegisterOperation("getPetById", func(ctx context.Context, req runtime.Request) (*runtime.Response, error) {
        params := GetPetByIDRequest{
            PetID: req.PathParams["petId"],
        }
        return client.GetPetByIDRequestCall(ctx, params)
    })
}
```

但这还需要把 `runtime.Request` 反序列化成 typed request struct，当前没有完整实现。

---

## 11. 测试覆盖评审

### 11.1 已有优点

- `pkg/generator` 测试较充分；
- `pkg/runtime/client_test.go` 覆盖了 path resolution、config、auth、error conversion；
- `go test -race ./...` 通过。

### 11.2 缺口

缺少测试覆盖：

| 文件 | 问题 |
|---|---|
| `pkg/runtime/converter.go` | JS/Go 转换是核心路径，但没有测试 |
| `pkg/runtime/exports.go` | WASM 入口函数没有测试 |
| `pkg/runtime/promise.go` | Promise resolve/reject/recover 没有测试 |
| `pkg/runtime/validator.go` | email/uuid/datetime 没有测试 |
| `cmd/generator/main.go` | CLI flag、JSON output、exit code 没有测试 |
| `pkg/generator/templates/main.go.tmpl` | build tag/import 没有测试 |
| `pkg/generator/templates/go.mod.tmpl` | module path 没有测试 |

建议增加：

1. CLI golden tests：
   - `--help`
   - `generate --dry-run`
   - `generate --output json`
   - invalid `--compiler`
   - `--wasm=false`

2. 生成产物测试：
   - 用 petstore spec 生成到 temp dir；
   - 校验 `go.mod` module path；
   - 校验 `cmd/wasm/main.go` 有 build tag；
   - 校验 `RuntimeImport` 被使用。

3. runtime validator tests：
   - valid/invalid email；
   - valid/invalid UUID；
   - valid/invalid RFC3339 datetime。

4. converter tests：
   - dangerous keys filtered；
   - nested object dangerous keys filtered；
   - unsupported type behavior；
   - negative number to uint behavior。

---

## 12. 关键问题重构前后对比

### 12.1 `--module` 不生效

#### 修改前

```go
if outDir != "" {
    cfg.OutputModule = outDir
}
```

```go
module {{.Config.OutputModule}}
```

#### 修改后

```go
type Config struct {
    ModuleName string
    OutputDir  string
}
```

```go
if outDir != "" {
    cfg.OutputDir = outDir
}
```

```go
module {{.Config.ModuleName}}
```

### 12.2 生成 `cmd/wasm/main.go` 缺少 build tag 且硬编码 import

#### 修改前

```go
package main

import (
    "github.com/fred29910/gowasm/pkg/runtime"
)
```

#### 修改后

```go
//go:build js && wasm

package main

import runtime "{{.Config.RuntimeImport}}"
```

### 12.3 required 字段零值校验错误

#### 修改前

```go
type GetPetByIDRequest struct {
    PetID int64 `json:"petId"`
}

func (r GetPetByIDRequest) Validate() error {
    if r.PetID == int64(0) {
        return fmt.Errorf("petId is required")
    }
    return nil
}
```

#### 修改后

```go
type GetPetByIDRequest struct {
    PetID *int64 `json:"petId"`
}

func (r GetPetByIDRequest) Validate() error {
    if r.PetID == nil {
        return fmt.Errorf("petId is required")
    }
    return nil
}
```

### 12.4 `GetConfig()` 返回内部指针

#### 修改前

```go
func (c *HTTPClient) GetConfig() *ClientConfig {
    return c.config
}
```

#### 修改后

```go
func (c *HTTPClient) GetConfig() *ClientConfig {
    c.mu.RLock()
    defer c.mu.RUnlock()

    headers := make(map[string]string, len(c.config.Headers))
    for k, v := range c.config.Headers {
        headers[k] = v
    }

    return &ClientConfig{
        BaseURL:     c.config.BaseURL,
        Timeout:     c.config.Timeout,
        Headers:     headers,
        Credentials: c.config.Credentials,
    }
}
```

### 12.5 path param 非法时静默失败

#### 修改前

```go
func safePathParam(v string) string {
    if strings.Contains(v, "..") {
        return ""
    }
    return v
}
```

#### 修改后

```go
func safePathParam(v string) error {
    unescaped, err := url.PathUnescape(v)
    if err != nil {
        return fmt.Errorf("decode path parameter: %w", err)
    }

    if strings.Contains(unescaped, "..") ||
        strings.Contains(unescaped, "//") ||
        strings.HasPrefix(unescaped, "/") {
        return fmt.Errorf("%w %q", ErrInvalidPathParam, v)
    }

    return nil
}
```

### 12.6 datetime validator 手写解析脆弱

#### 修改前

```go
func IsValidDateTime(dt string) bool {
    // 大量手写长度和字符判断
}
```

#### 修改后

```go
func IsValidDateTime(dt string) bool {
    _, err := time.Parse(time.RFC3339, dt)
    return err == nil
}
```

### 12.7 TS SDK 加载 Go WASM 不完整

#### 修改前

```ts
const wasmModule = await WebAssembly.instantiate(bytes, {
    env: {
        // Go WASM imports
    }
});
```

#### 修改后

```ts
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch(this.wasmUrl),
    go.importObject,
);
go.run(result.instance);
```

同时文档必须说明：

```text
main.wasm 必须与 wasm_exec.js 一起部署。
```

---

## 13. 优先级路线图

### P0：建议立即修复

1. 修复 `--module` 不生效，生成正确 `go.mod`。
2. 修复生成 `cmd/wasm/main.go` 的 build tag 和 runtime import。
3. required 字段改为 pointer 校验，避免 `0` / `false` 误判。
4. `GetConfig()` 返回 copy，避免 data race。
5. path param 非法时返回 error，而不是空字符串。
6. JSON 模式移除 `os.Exit(1)`，统一错误处理。

### P1：下一轮重构

1. 拆分 `pkg/runtime`。
2. 拆分 `cmd/generator/main.go`。
3. 增加 CLI tests 和生成产物 tests。
4. 修复 TS `query` 多值类型。
5. 修复 TS `load()` 的 Go WASM runtime 加载。
6. 为 build/tidy 增加 timeout/context。
7. 为 `NewHTTPClient` 初始化 nil `Headers`。

### P2：中长期增强

1. 引入 `log/slog`。
2. 支持 OpenAPI `allOf` / `oneOf` / `anyOf` / discriminator。
3. 支持 external `$ref`。
4. 支持 request/response interceptor。
5. 支持 retry/backoff。
6. 支持 shell completion。
7. 将 oxlint 从 generate 默认流程中拆出。

---

## 14. 总体评分

| 维度 | 评分 | 说明 |
|---|---:|---|
| CLI 交互设计 | 3/5 | flags 基本完整，但 examples、version 命令、默认 wasm、JSON exit 处理有问题 |
| 功能逻辑 | 3/5 | runtime 有安全保护，但 required 校验、path param、context cancellation 有缺陷 |
| Go 代码质量 | 3/5 | 测试不错，但 god package 和单文件大 Action 明显 |
| 错误处理 | 3.5/5 | `WASMError` 设计不错，但 Promise 和 CLI JSON 模式需加强 |
| 目录结构 | 3/5 | 有 `cmd/pkg/version`，但边界不清 |
| 生成 SDK 质量 | 2.5/5 | 有类型生成雏形，但 go.mod、main.go、required、TS load 有现成问题 |
| 依赖选择 | 4/5 | 总体克制，`norunners/vert` 需关注维护 |
| 测试覆盖 | 3/5 | generator 测试充分，runtime 关键 JS 层和 CLI 缺测试 |

---

## 15. 最终建议

这个项目已经具备可用原型和不错的测试基础，但如果目标是作为生产级 CLI/SDK generator，建议优先做三件事：

1. **先修生成产物正确性**：module path、main.go build tag、runtime import、required validation。
2. **再修 CLI 工程边界**：移除 `os.Exit`、拆 `main.go`、统一 JSON error、调整 `--wasm` 默认行为。
3. **最后做 runtime 拆分和高级 OpenAPI 支持**：否则后续功能会继续堆在 `pkg/runtime` 和 `cmd/generator/main.go` 里。

当前最应该马上修的是：

```text
--module 不生效
生成 cmd/wasm/main.go 缺少 js&&wasm build tag
required 字段零值校验错误
GetConfig 返回内部指针
path param 非法静默失败
```
