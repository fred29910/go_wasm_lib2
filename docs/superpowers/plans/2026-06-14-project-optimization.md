---
change: project-optimization
design-doc: docs/superpowers/specs/2026-06-14-project-optimization-design.md
base-ref: 1bb544d421a079a1cb5fc838a071f77ffcd0188d
---

# Implementation Plan

## 任务概览

共 10 组任务，按依赖顺序执行。

## 实施顺序

### Phase 1: 核心类型系统改进

- [ ] 1.1 在 `pkg/generator/openapi.go` 的 Schema 类型中添加 `Enum []interface{}` 字段
- [ ] 1.2 更新 `pkg/generator/types.go` 中的 `ToGoName` 函数，添加首字母缩写映射表（ID, URL, HTTP, JSON, API, UUID, JWT, HTML, XML, SQL, REST, gRPC）
- [ ] 1.3 更新 `pkg/generator/generator.go` 中的 `goType` 和 `tsType` 函数，识别 enum 值并生成对应类型
- [ ] 1.4 为 enum 类型生成独立的 TypeScript 类型定义

### Phase 2: 多响应支持

- [ ] 2.1 扩展 `GenerationModel` 和 `GeneratedOperation` 以包含所有响应码
- [ ] 2.2 更新 `buildOperations` 解析所有响应码（4xx, 5xx）的 schema
- [ ] 2.3 更新 Go 模板 (`sdk.go.tmpl`) 生成带前缀的响应结构体（如 `CreatePet200Response`, `CreatePet400Response`）
- [ ] 2.4 更新 TypeScript 模板 (`sdk.ts.tmpl`) 生成带前缀的响应接口

### Phase 3: 请求验证

- [ ] 3.1 在 Go 模板中添加 `Validate() error` 方法生成逻辑
- [ ] 3.2 实现必填字段非零值检查
- [ ] 3.3 实现 enum 值白名单校验
- [ ] 3.4 添加 `--validation` CLI flag 控制是否生成验证代码（默认开启）

### Phase 4: 自定义模板支持

- [ ] 4.1 在 `cmd/generator/main.go` 添加 `--go-template` 和 `--ts-template` CLI flags
- [ ] 4.2 更新 `Generator` 优先从文件加载模板，失败回退到嵌入模板
- [ ] 4.3 在 README 中记录所有可用模板变量

### Phase 5: CLI 增强

- [ ] 5.1 添加 `--dry-run` flag：解析 spec、构建模型、渲染模板但不写文件，输出文件列表
- [ ] 5.2 添加 `--verbose` flag：显示各阶段耗时、生成文件数量
- [ ] 5.3 添加 `--output json` flag：结构化输出便于 CI/CD 解析
- [ ] 5.4 改进错误消息，包含文件、行号、建议

### Phase 6: Oxlint 配置更新

- [ ] 6.1 更新根目录 `oxlintrc.json` 添加规则检测生成代码常见问题
- [ ] 6.2 同步更新 `cmd/generator/oxlintrc.json`

### Phase 7: 测试

- [ ] 7.1 在 `pkg/generator/openapi_test.go` 添加 enum 解析测试
- [ ] 7.2 在 `pkg/generator/types_test.go` 添加 ToGoName 首字母缩写测试
- [ ] 7.3 在 `pkg/generator/generator_test.go` 添加多响应生成测试
- [ ] 7.4 添加验证代码生成测试
- [ ] 7.5 使用 petstore 示例进行集成测试

### Phase 8: 文档更新

- [ ] 8.1 更新 README 添加新功能说明（enum、validation、custom templates）
- [ ] 8.2 添加 CLI 参考文档
- [ ] 8.3 添加命名变更迁移指南

### Phase 9: 验证

- [ ] 9.1 运行完整测试套件 (`go test ./...`)
- [ ] 9.2 生成 petstore SDK 并验证输出
- [ ] 9.3 运行 oxlint 检查生成的 TypeScript
- [ ] 9.4 测试 WASM 构建
- [ ] 9.5 测试 dry-run 和 verbose 模式

### Phase 10: 代码审查与收尾

- [ ] 10.1 代码审查
- [ ] 10.2 更新 oxlintrc.json 配置
- [ ] 10.3 提交所有更改