# CLI 参考文档

## 命令概览

```
gowasm-generator <command> [options]
```

| 命令 | 用途 |
|------|------|
| `generate` | 从 OpenAPI 规范生成 SDK |
| `init` | 创建示例项目结构 |
| `version` | 显示版本信息 |

## generate 命令

从 OpenAPI 3.x 规范生成 Go 和 TypeScript SDK。

### 用法

```bash
gowasm-generator generate [flags]
```

### 标志

#### 必需标志

| 标志 | 别名 | 类型 | 说明 |
|------|------|------|------|
| `--spec` | `-s` | string | OpenAPI 规范文件路径 (YAML) |

#### 输出选项

| 标志 | 别名 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--out` | `-o` | string | `./generated` | 输出目录路径 |
| `--output` | | string | `text` | 输出格式: `text` 或 `json` |

#### Go 代码生成选项

| 标志 | 别名 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--module` | `-m` | string | `github.com/fred29910/gowasm` | Go 模块名 |
| `--package` | `-p` | string | `generated` | Go 包名 |
| `--go-template` | | string | 内置模板 | 自定义 Go 模板文件路径 |

#### TypeScript 代码生成选项

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--ts-template` | string | 内置模板 | 自定义 TypeScript 模板文件路径 |

#### WASM 编译选项

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--wasm` | bool | `true` | 生成后是否编译 WASM |
| `--wasm-out` | string | `<out>/main.wasm` | WASM 输出路径 |
| `--compiler` | string | `auto` | 编译器选择: `auto`, `tinygo`, `go` |

#### 验证选项

| 标志 | 别名 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--validation` | `-v` | bool | `true` | 是否生成验证方法 |

#### Lint 选项

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--oxlintrc` | string | 内置配置 | oxlint 配置文件路径 |
| `--oxlint-disable` | bool | `false` | 是否禁用 oxlint |

#### 其他选项

| 标志 | 别名 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--dry-run` | | bool | `false` | 仅显示将要生成的文件，不实际写入 |
| `--verbose` | `-V` | bool | `false` | 显示详细进度信息 |

### 编译器选择 (`--compiler`)

| 值 | 行为 |
|----|------|
| `auto` | 自动检测 `tinygo`，如果存在则使用，否则回退到 `go` |
| `tinygo` | 强制使用 `tinygo`（未安装时报错） |
| `go` | 使用标准 Go 编译器 (`GOOS=js GOARCH=wasm`) |

### 输出格式 (`--output`)

#### text 格式 (默认)

```
Generated SDK in ./generated
```

启用 `--verbose` 时：

```
Parsing spec...
Building model: 3 schemas, 5 operations...
Rendering Go template...
Writing go.mod...
Writing main.go...
Rendering TypeScript template...
Writing files...
Generated 5 operations, 3 schemas, 5 files
All done!
```

#### json 格式

成功时：

```json
{
  "success": true,
  "files": [
    { "path": "generated.go", "size": 4523 },
    { "path": "go.mod", "size": 156 },
    { "path": "main.go", "size": 89 },
    { "path": "sdk.ts", "size": 3421 },
    { "path": "index.html", "size": 2156 }
  ],
  "totalSize": 10345,
  "duration": 1250
}
```

失败时：

```json
{
  "success": false,
  "duration": 50,
  "error": "stat openapi.yaml: no such file or directory"
}
```

### 生成文件结构

```
<output>/
├── generated.go     # Go 客户端代码：schema 结构体、请求/响应类型、验证方法、辅助函数
├── go.mod           # Go 模块定义
├── main.go          # WASM 入口文件
├── sdk.ts           # TypeScript SDK：接口定义、WASMSDK 类、类型化 API 函数
├── index.html       # 交互式演示页面（Tailwind CSS）
└── main.wasm        # 编译后的 WASM 二进制（--wasm=true 时）
```

### 示例

#### 基本生成

```bash
gowasm-generator generate -s api/openapi.yaml -o ./output
```

#### 指定模块和包名

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./sdk \
  -m github.com/mycompany/api \
  -p client
```

#### 使用 TinyGo 编译器

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./output \
  --compiler tinygo
```

#### 使用自定义模板

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./output \
  --go-template ./templates/custom.go.tmpl \
  --ts-template ./templates/custom.ts.tmpl
```

#### 仅生成代码，不编译 WASM

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./output \
  --wasm=false
```

#### 试运行（不写入文件）

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./output \
  --dry-run
```

输出：

```
Dry run: files that would be generated:
  generated.go (4523 bytes)
  go.mod (156 bytes)
  sdk.ts (3421 bytes)
  index.html (2156 bytes)

Total: 4 file(s), 10256 bytes
```

#### JSON 输出（用于 CI/CD）

```bash
gowasm-generator generate \
  -s api/openapi.yaml \
  -o ./output \
  --output json
```

#### 完整选项示例

```bash
gowasm-generator generate \
  -s ./api/openapi.yaml \
  -o ./sdk/generated \
  -m github.com/myorg/myproject \
  -p apiclient \
  --compiler tinygo \
  --wasm-out ./build/sdk.wasm \
  --validation \
  --verbose \
  --output json
```

## init 命令

创建示例项目结构。

### 用法

```bash
gowasm-generator init
```

### 输出

```
Creating sample project structure...
Created directories: specs/, generated/, build/
Next: place your spec in specs/ and run:
  gowasm-generator generate -s specs/openapi.yaml -o generated/
```

### 创建的结构

```
.
├── specs/       # 放置 OpenAPI 规范
├── generated/   # 生成的 SDK 输出
└── build/       # WASM 构建产物
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `SPEC` | Makefile 中的 OpenAPI 规范路径 |
| `OUT` | Makefile 中的输出目录 |

## 退出码

| 码 | 含义 |
|----|------|
| 0 | 成功 |
| 1 | 错误（参数无效、文件未找到、编译失败等） |

## Makefile 快捷命令

Makefile 提供了以下快捷命令：

```bash
make build              # 构建 WASM (标准 Go)
make build-tinygo       # 构建 WASM (TinyGo)
make build-all          # 构建所有版本
make generate SPEC=... OUT=...  # 生成 SDK
make dev-generate       # 生成 Petstore 示例
make test               # 运行测试
make test-compile       # 测试编译
make verify             # 完整验证
make clean              # 清理构建产物
make lint-ts            # 对生成的 TS 代码运行 oxlint 检查
make lint-ts-fix        # 对生成的 TS 代码运行 oxlint 并自动修复
```

### Makefile 示例

```bash
# 生成 SDK
make generate SPEC=./api/openapi.yaml OUT=./output

# Petstore 示例
make dev-generate

# 对生成的代码进行 TS 检查
make lint-ts OUT=./output

# 自动修复 TS 检查问题
make lint-ts-fix OUT=./output
```

## Task 命令

Taskfile.yml 提供了跨平台支持：

```bash
task build              # 构建 WASM (标准 Go)
task build:tinygo       # 构建 WASM (TinyGo)
task build:all          # 构建所有版本
task generate SPEC=... OUT=...  # 生成 SDK
task dev:generate       # 生成 Petstore 示例
task test               # 运行测试
task test:compile       # 测试编译
task verify             # 完整验证
task clean              # 清理构建产物
task lint-ts            # 对生成的 TS 代码运行 oxlint 检查
task lint-ts-fix        # 对生成的 TS 代码运行 oxlint 并自动修复
```
