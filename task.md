# 任务完成状态

基于 plan2.md，截至 2026-05-15。

## 已完成

- Milestone 1: Callouts 支持 ✅
- Milestone 2: Properties / Aliases 增强 ✅
- Milestone 3: 正文标签与标签树 ✅（补全了标签树）
- Milestone 4: 块引用 Block Reference ✅
- Milestone 5: Mermaid 支持 ✅
- Milestone 6: JSON Canvas 只读预览 ✅（补全了 Scanner/API/前端 Viewer）
- Milestone 7: Graph View 图谱 ✅（新建）
- Milestone 8: Dashboard 首页 ✅（新建）
- Milestone 9: Dataview Lite / Vault Query ✅（新建）
- Milestone 10: 测试与文档 ✅（补全了 callout_test、indexer_test、README、CHANGELOG）

## 待改进项

- [ ] Mermaid: CDN 替换为本地 `web/vendor/mermaid.min.js`（离线/内网场景）
- [ ] Block Reference: 补充 `/api/block` HTTP 端点（indexer 方法已有）
- [ ] 暗色模式全面样式检查（Canvas、Graph、Dashboard、Vault Query）
- [ ] `callout_test.go` 扩展：更多边界用例
- [ ] Indexer 测试：alias 冲突场景
