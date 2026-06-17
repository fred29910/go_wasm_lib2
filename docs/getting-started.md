# 快速开始指南

## 前置要求

| 工具 | 版本 | 用途 | 必需 |
|------|------|------|------|
| Go | 1.25.1+ | 编译 Go 代码和 WASM | ✅ |
| TinyGo | 0.41.1+ | 生成更小的 WASM 二进制 | ❌ |
| Node.js | 18+ | 运行 oxlint 检查生成的 TS 代码 | ❌ |
| Task | 可选 | 替代 Make 的构建工具 | ❌ |

## 安装

### 1. 克隆仓库

```bash
git clone https://github.com/fred29910/gowasm.git
cd gowasm
```

### 2. 安装依赖

```bash
# 下载 Go 依赖
go mod tidy
go mod download

# 安装 TinyGo (可选，推荐)
# macOS
brew install tinygo

# Linux (TinyGo 0.41.1)
wget https://github.com/tinygo-org/tinygo/releases/download/v0.41.1/tinygo_0.41.1_amd64.deb
sudo dpkg -i tinygo_0.41.1_amd64.deb
```

## 构建 WASM 运行时

### 使用标准 Go 编译器

```bash
make build
# 或
task build
```

输出: `build/main.wasm` (约 2-5 MB)

### 使用 TinyGo 编译器

```bash
make build-tinygo
# 或
task build:tinygo
```

输出: `build/tinymain.wasm` (约 200-500 KB)

### 构建所有版本

```bash
make build-all
# 或
task build:all
```

## 生成 SDK

### 基本用法

```bash
# 使用 Make
make generate SPEC=path/to/openapi.yaml OUT=./output

# 使用 CLI 直接运行
go run ./cmd/generator generate -s path/to/openapi.yaml -o ./output
```

### 使用 Petstore 示例

```bash
make dev-generate
```

生成产物位于 `examples/petstore/generated/`:
- `generated.go` - Go 客户端代码
- `go.mod` - Go 模块定义
- `main.go` - WASM 入口
- `sdk.ts` - TypeScript SDK
- `index.html` - 演示页面

### 完整选项

```bash
gowasm-generator generate \
  -s ./api/openapi.yaml \        # OpenAPI 规范路径 (必需)
  -o ./sdk/generated \           # 输出目录 (默认: ./generated)
  -m github.com/myorg/myproject \ # Go 模块名
  -p apiclient \                 # Go 包名 (默认: generated)
  --compiler tinygo \            # 编译器: auto|tinygo|go (默认: auto)
  --wasm-out ./build/sdk.wasm \  # WASM 输出路径
  --validation \                 # 生成验证方法 (默认: true)
  --verbose                      # 显示详细进度
```

> **注意**：`--wasm` 默认为 `false`，如需编译 WASM 需显式指定 `--wasm`。

## 在浏览器中使用

### 1. 加载 WASM

```html
<!DOCTYPE html>
<html>
<head>
  <title>WASM API Client</title>
</head>
<body>
  <script type="module">
    import { WASMSDK } from './sdk.js';

    const sdk = new WASMSDK('./main.wasm');

    // 加载 WASM 模块（自动加载 wasm_exec.js 并等待就绪）
    await sdk.load();

    // 初始化客户端
    await sdk.init({
      baseUrl: 'https://api.example.com',
      timeout: 30,
      headers: { 'Content-Type': 'application/json' }
    });

    // 设置认证令牌
    sdk.setAuthToken('your-jwt-token', 'Bearer');

    // 调用 API
    const response = await sdk.callAPI('getUsers', {
      method: 'GET',
      path: '/users',
      query: { page: '1', limit: '10' }
    });

    console.log('Status:', response.status);
    console.log('Body:', response.body);
  </script>
</body>
</html>
```

### 2. 使用生成的类型化方法

```typescript
import { WASMSDK, createPet, getPetById } from './sdk.js';

const sdk = new WASMSDK('./main.wasm');
await sdk.load();
await sdk.init({ baseUrl: 'https://petstore3.swagger.io/api/v3' });

// 使用类型化方法
const response = await createPet({
  body: { name: 'Fluffy', status: 'available' }
});

const pet = response.body;
console.log('Created pet:', pet.name);
```

## 代码检查 (oxlint)

生成 SDK 后，可以使用 oxlint 检查生成的 TypeScript 代码质量：

```bash
# 检查生成的 TS 代码
make lint-ts OUT=./output

# 检查并自动修复问题
make lint-ts-fix OUT=./output
```

oxlint 配置文件位于根目录 `oxlintrc.json`，可通过 `--oxlintrc` 标志使用自定义配置。

## 自定义模板

### 快速开始

```bash
# 使用自定义 Go 模板
gowasm-generator generate -s openapi.yaml -o ./output --go-template ./my-go.tmpl

# 使用自定义 TypeScript 模板
gowasm-generator generate -s openapi.yaml -o ./output --ts-template ./my-tmpl

# 同时使用两个自定义模板
gowasm-generator generate -s openapi.yaml -o ./output \
  --go-template ./my-go.tmpl \
  --ts-template ./my-ts.tmpl
```

### 示例模板

`examples/templates/` 目录包含现成的自定义模板示例：

| 模板 | 说明 |
|------|------|
| `custom.go.tmpl` | 带分区注释和详细字段文档的 Go 模板 |
| `custom.ts.tmpl` | 带 JSDoc 注释和类型化 API 函数的 TS 模板 |

> 📖 模板变量和详细用法请参阅 [生成器 API 文档](./generator-api.md)。

## 运行测试

```bash
# 运行 Go 单元测试
make test
# 或
task test

# 测试 WASM 编译
make test-compile

# 完整验证 (deps, build, test, generate)
make verify
```

## 项目初始化

创建新的项目结构：

```bash
gowasm-generator init
```

创建以下目录结构：
```
.
├── specs/       # 放置 OpenAPI 规范
├── generated/   # 生成的 SDK 输出
└── build/       # WASM 构建产物
```

## 构建系统对比

| 特性 | Make | Task |
|------|------|------|
| 语法 | Tab-based, shell 命令 | YAML-based, 跨平台 |
| 依赖 | 需要 `make` | 需要 `task` |
| 源跟踪 | 有限 | 内置 (`sources`, `generates`) |
| 跨平台 | Unix 为主 | 完整 Windows 支持 |
| 详细输出 | 默认 | 可配置 |

## 下一步

- 阅读 [CLI 参考文档](./cli-reference.md) 了解所有命令选项
- 阅读 [运行时 API 文档](./runtime-api.md) 了解 WASM 导出函数和安全特性
- 阅读 [生成器 API 文档](./generator-api.md) 了解模板系统和自定义模板变量
- 查看 [示例](../examples/petstore/) 了解完整用法
- 阅读 [架构文档](./architecture.md) 了解系统设计