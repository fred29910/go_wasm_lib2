---
change: project-optimization
design-doc: docs/superpowers/specs/2026-06-14-project-optimization-design.md
base-ref: 05d4b073ee5d7d529f8d5dd6ee4ce5b7905a7a34
---

# Project Optimization 恢复实施计划

## 当前状态

- OpenSpec artifacts 已完成：proposal、design、specs、tasks 均为 done。
- Comet 当前处于 build 阶段，`build_mode: subagent-driven-development`，`tdd_mode: tdd`，`isolation: branch`。
- 本计划用于恢复执行，只保留 `openspec/changes/project-optimization/tasks.md` 中尚未完成的实现、文档和验证任务。
- 工作区存在与当前 change 目标无关的 `docs/reviews/*` 删除改动；本计划不处理这些文件，也不得把它们纳入当前 change 提交。

## 剩余任务

### Phase 8: 测试收尾

- [x] 8.5 Add integration test with petstore example

### Phase 9: 文档更新

- [x] 9.1 Update README with new features (enums, validation, custom templates)
- [x] 9.2 Add CLI reference for new flags
- [x] 9.3 Add migration guide for naming convention changes
- [x] 9.4 Update examples if needed

### Phase 10: 验证

- [ ] 10.1 Run full test suite (go test ./...)
- [ ] 10.2 Generate petstore SDK and verify output
- [ ] 10.3 Run oxlint on generated TypeScript
- [ ] 10.4 Test WASM build with generated code
- [ ] 10.5 Test dry-run and verbose modes
- [ ] 10.6 Test custom template override

## 执行顺序

1. 先完成 `8.5 Add integration test with petstore example`，确认 petstore 示例生成路径和断言范围。
2. 更新 README、CLI reference、命名迁移说明和示例，确保文档与已实现能力一致。
3. 执行全量验证：`go test ./...`、petstore SDK 生成、TypeScript oxlint、WASM build、dry-run/verbose/custom template 验证。
4. 每个任务完成后由主会话同步勾选 OpenSpec task，并保留 subagent 审查证据。
