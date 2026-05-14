下面是一份可以直接交给 **Claude Code** 执行的任务计划书，目标是在已有 Go 版只读 Obsidian Vault Web Reader 基础上，继续补齐 **Obsidian 特色能力**：Callouts、Properties、Aliases、块引用、Canvas、Mermaid、图谱、Dataview Lite 等。

---

# Go Obsidian Vault Web Reader 增强任务计划书

## 0. 项目背景

当前项目目标是开发一个：

```text
Go 实现
只读
低资源占用
运行在 Linux 服务器
通过浏览器查看 Obsidian Vault
高兼容 Obsidian 双链和特色语法
```

已有基础能力假设包括：

```text
1. Vault 扫描
2. Markdown 渲染
3. 目录树
4. [[双链]]
5. 附件预览
6. SQLite 索引
7. 反链
8. 搜索
9. Docker 部署
```

本任务计划继续增强 Obsidian 特色能力。

---

# 1. 总体开发原则

Claude Code 执行时必须遵守：

```text
1. Vault 必须只读
2. 不修改用户原始 Markdown 文件
3. 不往 Vault 写缓存文件
4. 所有索引、收藏、访问历史写入 DATA_DIR
5. 所有路径访问必须做 realpath 校验
6. 禁止路径穿越
7. 支持中文文件名
8. 支持空格路径
9. 不引入 Node 构建链
10. 前端优先使用 HTML + CSS + Vanilla JS
11. Go 后端保持单进程运行
12. 每个阶段完成后必须补充测试样例
```

---

# 2. 推荐执行顺序

```text
Milestone 1：Callouts 支持
Milestone 2：Properties / Aliases 增强
Milestone 3：正文标签与标签树
Milestone 4：块引用 Block Reference
Milestone 5：Mermaid 支持
Milestone 6：JSON Canvas 只读预览
Milestone 7：Graph View 图谱
Milestone 8：Dashboard 首页
Milestone 9：Dataview Lite / Vault Query
Milestone 10：打磨、测试、文档
```

---

# 3. Milestone 1：Obsidian Callouts 支持

## 3.1 目标

支持 Obsidian 常见 Callout 语法。

输入：

```md
> [!note]
> 这是一条普通提示

> [!warning] 注意
> 这里是警告内容

> [!tip]- 折叠提示
> 默认折叠内容

> [!example]+ 展开示例
> 默认展开内容
```

输出为带样式的 HTML 卡片。

---

## 3.2 支持类型

```text
note
abstract
summary
tldr
info
todo
tip
hint
important
success
check
done
question
help
faq
warning
caution
attention
failure
fail
missing
danger
error
bug
example
quote
cite
```

---

## 3.3 开发任务

```text
1. 在 parser 模块新增 callout.go
2. 实现 DetectCallouts(markdown string) 或 Markdown AST 扩展
3. 识别 blockquote 第一行中的 [!type]
4. 解析标题
5. 解析折叠状态：
   - [!tip]- 默认折叠
   - [!tip]+ 默认展开
6. 渲染为 div.callout
7. CSS 支持不同类型颜色和图标
8. 支持内部 Markdown 渲染
9. 给 test-vault 增加 callout 测试文件
```

---

## 3.4 建议文件

```text
internal/parser/callout.go
internal/parser/callout_test.go
web/style.css
test-vault/10_Reference/Callouts 测试.md
```

---

## 3.5 验收标准

```text
1. note/warning/tip/example 正常渲染
2. 标题正常显示
3. 折叠状态可点击展开/收起
4. Callout 内部代码块、列表、链接不丢失
5. 暗色模式下样式可读
```

---

# 4. Milestone 2：Properties / Aliases 增强

## 4.1 目标

把 frontmatter 从“原始 YAML 解析”升级成 Obsidian Properties 系统。

支持：

```yaml
---
title: OpenClaw 配置排查
aliases:
  - OpenClaw Gateway
  - OpenClaw 网关
tags:
  - openclaw
  - debug
status: active
type: debug-note
source: https://example.com
created: 2026-05-14
updated: 2026-05-14
---
```

---

## 4.2 开发任务

```text
1. 新增 properties 表
2. 解析 YAML frontmatter 中所有 key/value
3. 标准化 aliases、tags、created、updated、status、type、source
4. 右侧栏显示 Properties
5. aliases 加入链接解析索引
6. [[OpenClaw Gateway]] 能解析到设置了 alias 的真实文件
7. 支持属性筛选 API
```

---

## 4.3 数据库表

```sql
CREATE TABLE IF NOT EXISTS properties (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  key TEXT NOT NULL,
  value TEXT,
  value_type TEXT
);

CREATE INDEX IF NOT EXISTS idx_properties_key ON properties(key);
CREATE INDEX IF NOT EXISTS idx_properties_file_path ON properties(file_path);
```

---

## 4.4 API

```text
GET /api/properties?path=xxx
GET /api/filter?key=status&value=active
GET /api/filter?key=type&value=debug-note
```

---

## 4.5 建议文件

```text
internal/parser/properties.go
internal/indexer/properties.go
internal/resolver/alias.go
internal/server/handlers_properties.go
web/app.js
web/style.css
```

---

## 4.6 验收标准

```text
1. 笔记右侧显示 Properties
2. aliases 不显示为普通字段，而是作为别名区块显示
3. tags 正常进入标签索引
4. [[别名]] 可以跳转到真实文件
5. 多个文件拥有相同 alias 时显示冲突候选
```

---

# 5. Milestone 3：正文标签与标签树

## 5.1 目标

除了 YAML tags，还要支持正文标签：

```md
#openclaw
#debug/proxy
#codex/oauth
```

---

## 5.2 开发任务

```text
1. 在 Markdown 正文中提取 inline tags
2. 排除代码块内的 #xxx
3. 支持嵌套标签 debug/proxy
4. 更新 tags 表
5. 新增标签树 API
6. 前端增加标签页
7. 正文标签渲染为可点击链接
```

---

## 5.3 API

```text
GET /api/tags
GET /api/tag?name=openclaw
GET /api/tag-tree
```

---

## 5.4 标签树返回示例

```json
{
  "items": [
    {
      "name": "debug",
      "count": 10,
      "children": [
        {
          "name": "proxy",
          "fullName": "debug/proxy",
          "count": 4
        }
      ]
    }
  ]
}
```

---

## 5.5 建议文件

```text
internal/parser/tags.go
internal/indexer/tags.go
internal/server/handlers_tags.go
web/app.js
web/style.css
```

---

## 5.6 验收标准

```text
1. YAML tags 和正文 tags 都能识别
2. 代码块里的 #tag 不被误识别
3. #debug/proxy 显示为层级标签
4. 点击标签可以查看相关笔记
```

---

# 6. Milestone 4：块引用 Block Reference

## 6.1 目标

支持 Obsidian block id：

```md
这是一段重要结论。 ^abc123
```

支持链接：

```md
[[OpenClaw#^abc123]]
[[OpenClaw#^abc123|重要结论]]
```

---

## 6.2 开发任务

```text
1. 扫描 Markdown 中的 block id
2. 建立 block_id 到 file_path、line、text 的映射
3. 修改 wikilink 解析，支持 heading 为 ^xxx
4. 渲染 HTML 时给目标块加 id
5. 点击链接后滚动到目标块
6. 目标块短暂高亮
```

---

## 6.3 数据库表

```sql
CREATE TABLE IF NOT EXISTS blocks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  block_id TEXT NOT NULL,
  text TEXT,
  line_start INTEGER,
  line_end INTEGER
);

CREATE INDEX IF NOT EXISTS idx_blocks_file_path ON blocks(file_path);
CREATE INDEX IF NOT EXISTS idx_blocks_block_id ON blocks(block_id);
```

---

## 6.4 API

```text
GET /api/block?path=xxx&id=abc123
```

---

## 6.5 建议文件

```text
internal/parser/blocks.go
internal/indexer/blocks.go
internal/resolver/block.go
web/app.js
web/style.css
```

---

## 6.6 验收标准

```text
1. 文档中的 ^abc123 不作为普通正文显示得很突兀
2. [[File#^abc123]] 可以跳转
3. 页面自动滚动到目标块
4. 目标块高亮 2 秒
5. 不影响普通 heading 链接
```

---

# 7. Milestone 5：Mermaid 支持

## 7.1 目标

支持 Obsidian 中常见 Mermaid 代码块：

````md
```mermaid
graph TD
  A --> B
```
````

---

## 7.2 开发任务

```text
1. 后端渲染 Markdown 时保留 mermaid 代码块
2. 前端检测 code.language-mermaid
3. 加载 Mermaid 浏览器脚本
4. 渲染为 SVG
5. 支持暗色模式
6. 渲染失败时显示原始代码和错误提示
```

---

## 7.3 注意

第一版可以使用 CDN，但内网服务器建议支持本地静态文件。

建议：

```text
web/vendor/mermaid.min.js
```

---

## 7.4 建议文件

```text
web/vendor/mermaid.min.js
web/app.js
web/style.css
```

---

## 7.5 验收标准

```text
1. graph TD 正常渲染
2. sequenceDiagram 正常渲染
3. classDiagram 正常渲染
4. 渲染失败不导致整个页面白屏
```

---

# 8. Milestone 6：JSON Canvas 只读预览

## 8.1 目标

支持 Obsidian `.canvas` 文件只读查看。

Canvas 文件是 JSON，典型结构：

```json
{
  "nodes": [
    {
      "id": "node1",
      "type": "text",
      "text": "Hello",
      "x": 0,
      "y": 0,
      "width": 300,
      "height": 200
    }
  ],
  "edges": []
}
```

---

## 8.2 支持节点类型

```text
text
file
link
group
```

---

## 8.3 支持边

```text
fromNode
fromSide
toNode
toSide
label
color
```

---

## 8.4 开发任务

```text
1. scanner 识别 .canvas 文件
2. 新增 canvas parser
3. 定义 CanvasDocument、CanvasNode、CanvasEdge 结构
4. 实现 /api/canvas?path=xxx
5. 前端新增 canvas viewer
6. 根据 x/y/width/height 绝对定位节点
7. 使用 SVG 绘制边
8. 支持鼠标拖动画布
9. 支持滚轮缩放
10. file 节点点击后跳转到对应 note
11. link 节点点击打开外部链接
12. group 节点渲染为分组边框
```

---

## 8.5 Go 结构体

```go
type CanvasDocument struct {
    Nodes []CanvasNode `json:"nodes"`
    Edges []CanvasEdge `json:"edges"`
}

type CanvasNode struct {
    ID     string `json:"id"`
    Type   string `json:"type"`
    Text   string `json:"text,omitempty"`
    File   string `json:"file,omitempty"`
    URL    string `json:"url,omitempty"`
    X      int    `json:"x"`
    Y      int    `json:"y"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
    Color  string `json:"color,omitempty"`
    Label  string `json:"label,omitempty"`
}

type CanvasEdge struct {
    ID       string `json:"id"`
    FromNode string `json:"fromNode"`
    FromSide string `json:"fromSide"`
    ToNode   string `json:"toNode"`
    ToSide   string `json:"toSide"`
    Label    string `json:"label,omitempty"`
    Color    string `json:"color,omitempty"`
}
```

---

## 8.6 建议文件

```text
internal/parser/canvas.go
internal/server/handlers_canvas.go
web/canvas.js
web/style.css
test-vault/30_Dashboard/知识地图.canvas
```

---

## 8.7 验收标准

```text
1. .canvas 文件出现在目录树中
2. 点击 .canvas 打开画布视图
3. text 节点正常显示
4. file 节点正常显示标题并可跳转
5. link 节点可打开外链
6. group 节点有边框
7. edges 正常连线
8. 支持缩放和拖动
9. 大坐标节点不会导致页面错乱
```

---

# 9. Milestone 7：Graph View 图谱

## 9.1 目标

基于 links 表实现 Obsidian 风格图谱。

支持：

```text
1. 全局图谱
2. 当前笔记局部图谱
3. 按文件夹过滤
4. 按标签过滤
5. 节点点击跳转
```

---

## 9.2 API

```text
GET /api/graph
GET /api/graph?path=20_Debug/OpenClaw.md&depth=2
GET /api/graph?tag=openclaw
GET /api/graph?folder=20_Debug
```

---

## 9.3 返回结构

```json
{
  "nodes": [
    {
      "id": "20_Debug/OpenClaw.md",
      "title": "OpenClaw",
      "path": "20_Debug/OpenClaw.md",
      "group": "20_Debug"
    }
  ],
  "edges": [
    {
      "source": "20_Debug/Codex.md",
      "target": "20_Debug/OpenClaw.md"
    }
  ]
}
```

---

## 9.4 开发任务

```text
1. 基于 files 和 links 表生成 graph 数据
2. 过滤 unresolved links
3. 实现 /api/graph
4. 前端新增 graph 页面
5. 使用 SVG 或 Canvas 绘制 force graph
6. 支持搜索节点
7. 支持点击节点打开笔记
8. 支持局部图谱入口
```

---

## 9.5 建议文件

```text
internal/server/handlers_graph.go
internal/indexer/graph.go
web/graph.js
web/style.css
```

---

## 9.6 验收标准

```text
1. 能看到节点和连线
2. 点击节点打开对应笔记
3. 当前笔记右侧可以打开局部图谱
4. 大 Vault 下接口响应不会明显卡死
5. 可以限制最大节点数
```

---

# 10. Milestone 8：Dashboard 首页

## 10.1 目标

做一个适合知识库日常使用的首页。

针对目录：

```text
00_Inbox
10_Reference
20_Debug
30_Dashboard
90_Templates
```

首页展示：

```text
1. 最近修改
2. 00_Inbox 未整理
3. 最近 Debug 文档
4. 热门标签
5. status = active 的文档
6. type = debug-note 的文档
7. Canvas 入口
```

---

## 10.2 API

```text
GET /api/dashboard
```

---

## 10.3 返回结构

```json
{
  "recent": [],
  "inbox": [],
  "active": [],
  "debug": [],
  "tags": [],
  "canvas": []
}
```

---

## 10.4 开发任务

```text
1. 实现 dashboard 查询
2. 首页加载 /api/dashboard
3. 增加卡片式布局
4. 支持快速跳转
5. 支持按目录配置 Dashboard 区块
```

---

## 10.5 建议文件

```text
internal/server/handlers_dashboard.go
internal/indexer/dashboard.go
web/app.js
web/style.css
```

---

## 10.6 验收标准

```text
1. 打开首页不是空白
2. 能快速进入最近修改文档
3. 能看到 Inbox 内容
4. 能看到 Debug 文档
5. 能看到 Canvas 文件
```

---

# 11. Milestone 9：Dataview Lite / Vault Query

## 11.1 目标

不完整兼容 Dataview，而是实现可控的轻量查询块。

支持自定义代码块：

````md
```vault-query
type: table
from: 20_Debug
where:
  status: active
sort:
  updated: desc
limit: 20
fields:
  - title
  - status
  - updated
  - tags
```
````

---

## 11.2 暂不支持

```text
1. Dataview JS
2. 复杂表达式
3. 函数调用
4. 隐式字段完整兼容
5. Tasks 完整查询语法
```

---

## 11.3 开发任务

```text
1. 识别 vault-query 代码块
2. 解析 YAML 查询内容
3. 根据 files/properties/tags 表查询
4. 渲染为 table/list/card
5. 查询失败时显示错误块
6. 限制最大返回数量，防止页面过重
```

---

## 11.4 支持查询类型

```text
table
list
cards
```

---

## 11.5 建议文件

```text
internal/parser/vault_query.go
internal/query/query.go
internal/query/render.go
web/style.css
```

---

## 11.6 验收标准

```text
1. vault-query table 正常显示
2. from 文件夹过滤生效
3. where status 生效
4. sort updated desc 生效
5. limit 生效
6. 查询错误不会导致页面崩溃
```

---

# 12. Milestone 10：打磨、测试、文档

## 12.1 测试任务

```text
1. 为 wikilink parser 补充单元测试
2. 为 callout parser 补充单元测试
3. 为 block parser 补充单元测试
4. 为 canvas parser 补充单元测试
5. 为 path security 补充单元测试
6. 为 resolver 补充同名文件、alias 冲突测试
```

---

## 12.2 安全测试

必须测试：

```text
/assets?path=../../etc/passwd
/api/note?path=../../etc/passwd
/api/canvas?path=../../etc/passwd
```

结果必须是：

```text
403 Forbidden
```

---

## 12.3 README 更新

README 必须包含：

```text
1. 项目介绍
2. 功能清单
3. 不支持功能
4. 快速启动
5. Docker 部署
6. systemd 部署
7. Nginx Basic Auth
8. Vault 只读说明
9. Canvas 支持说明
10. Callouts 支持说明
11. Dataview Lite 说明
12. 常见问题
```

---

# 13. Claude Code 执行 Prompt

下面这段可以直接给 Claude Code：

````text
你正在维护一个 Go 实现的只读 Obsidian Vault Web Reader。

项目目标：
- Linux 服务器运行
- 读取已有 Obsidian Vault
- 浏览器查看 Markdown 知识库
- 不修改 Vault 原文件
- 单进程、低资源占用
- 高兼容 Obsidian 常见语法

当前基础能力假设已有：
- Vault 扫描
- Markdown 渲染
- 目录树
- [[双链]]
- 附件预览
- SQLite 索引
- 反链
- 搜索
- Docker 部署

请按以下 Milestone 顺序逐步开发，不要一次性大改：

Milestone 1：Callouts 支持
- 支持 > [!note]、> [!warning]、> [!tip]-、> [!example]+
- 支持标题、折叠状态、内部 Markdown
- 增加 CSS 样式和测试样例

Milestone 2：Properties / Aliases 增强
- 解析 YAML frontmatter 为 properties
- 支持 aliases、tags、status、type、source、created、updated
- aliases 参与 [[双链]] 解析
- 右侧栏显示 Properties

Milestone 3：正文标签与标签树
- 支持正文 #tag 和 #debug/proxy
- 排除代码块内标签
- 增加 /api/tag-tree
- 标签可点击跳转

Milestone 4：块引用 Block Reference
- 支持 ^blockid
- 支持 [[File#^blockid]]
- 点击后滚动并高亮目标块

Milestone 5：Mermaid 支持
- 支持 ```mermaid 代码块
- 前端渲染为图
- 渲染失败时展示错误，不允许白屏

Milestone 6：JSON Canvas 只读预览
- 支持 .canvas 文件
- 解析 JSON Canvas nodes/edges
- 支持 text/file/link/group 节点
- 支持 SVG 连线
- 支持缩放和拖动画布
- file 节点可跳转到笔记

Milestone 7：Graph View 图谱
- 基于 links 表生成图谱
- 支持全局图谱和局部图谱
- 节点点击打开笔记
- 支持最大节点数限制

Milestone 8：Dashboard 首页
- 展示最近修改、Inbox、Debug 文档、热门标签、Canvas 入口
- 实现 /api/dashboard

Milestone 9：Dataview Lite / Vault Query
- 不完整兼容 Dataview
- 支持 ```vault-query YAML 查询块
- 支持 table/list/cards
- 支持 from、where、sort、limit、fields

Milestone 10：测试与文档
- 补充 parser、resolver、path security、canvas 测试
- 更新 README
- 保证 Docker 部署仍可用

开发约束：
1. Vault 必须只读
2. 不允许向 Vault 写入文件
3. 所有状态写入 DATA_DIR 或 SQLite
4. 不引入 Node 构建链
5. 前端使用 HTML + CSS + Vanilla JS
6. 所有路径参数必须防止路径穿越
7. 支持中文文件名和空格路径
8. 每个 Milestone 完成后运行测试
9. 每个 Milestone 完成后更新 CHANGELOG
10. 不要重构无关模块

优先先完成 Milestone 1。
完成后输出：
- 修改了哪些文件
- 新增了哪些 API
- 如何测试
- 还有哪些已知限制
````

---

# 14. 分阶段执行命令建议

## 第一次给 Claude Code

```text
先执行 Milestone 1：Obsidian Callouts 支持。不要做其他 Milestone。
完成后运行 go test ./...，并更新 README 中的 Callouts 说明。
```

---

## 第二次给 Claude Code

```text
继续执行 Milestone 2：Properties / Aliases 增强。
要求 aliases 参与 wikilink resolver。
完成后补充 alias 冲突测试。
```

---

## 第三次给 Claude Code

```text
继续执行 Milestone 3 和 Milestone 4。
实现正文标签、标签树、块引用。
不要改动 Canvas 相关功能。
```

---

## 第四次给 Claude Code

```text
继续执行 Milestone 5：Mermaid 支持。
要求不引入 Node 构建链，前端使用本地 vendor 脚本。
渲染失败不得导致页面白屏。
```

---

## 第五次给 Claude Code

```text
继续执行 Milestone 6：JSON Canvas 只读预览。
优先实现 text、file、link、group 四种节点。
支持 SVG edges、缩放、拖拽。
不做 Canvas 编辑。
```

---

## 第六次给 Claude Code

```text
继续执行 Milestone 7 和 Milestone 8。
实现 Graph View 和 Dashboard 首页。
注意大 Vault 下要限制最大节点数。
```

---

## 第七次给 Claude Code

```text
继续执行 Milestone 9：Vault Query。
不要完整实现 Dataview，只做 vault-query YAML 查询块。
支持 table/list/cards。
```

---

## 第八次给 Claude Code

```text
执行 Milestone 10。
补测试、补 README、补部署文档、补 CHANGELOG。
重点测试路径穿越、中文文件名、alias 冲突、canvas 解析失败。
```

---

# 15. 最终版本功能清单

完成后应具备：

```text
基础能力：
- 目录树
- Markdown 阅读
- 搜索
- 反链
- 标签
- 附件预览

Obsidian 兼容：
- [[双链]]
- [[别名]]
- [[标题链接]]
- [[块引用]]
- ![[附件嵌入]]
- YAML Properties
- aliases
- tags
- Callouts
- Mermaid
- JSON Canvas
- Graph View

增强能力：
- Dashboard 首页
- 标签树
- 属性筛选
- Vault Query
- 只读安全部署
```

---

# 16. 最关键的取舍

不要让 Claude Code 一次性实现完整 Obsidian。

明确边界：

```text
做：
- 只读查看
- 常见语法兼容
- 快速搜索
- 知识网络展示
- Canvas 只读预览

不做：
- 在线编辑
- 插件系统
- Dataview 完整语言
- Canvas 编辑器
- 多人协同
- Obsidian Sync 替代品
```

最终目标一句话：

> **做一个轻量、只读、低资源、服务器端运行、重点兼容 Obsidian 双链与 Canvas 的 Web 知识库浏览器。**
