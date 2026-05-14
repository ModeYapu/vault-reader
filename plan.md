下面是一份可直接交给 AI Agent / 开发者执行的任务计划书。

# Go 只读 Obsidian Vault Web Reader 任务计划书

## 1. 项目定位

项目名称建议：

```text
Vault Reader
```

项目目标：

> 开发一个运行在 Linux 服务器上的只读 Obsidian Vault Web Reader，通过浏览器查看服务器已有的 Obsidian Markdown 知识库，重点兼容 Obsidian 双链、反链、附件、标签和搜索。

项目不做：

```text
不做在线编辑
不修改 Vault 原文件
不实现完整 Obsidian 插件系统
不做 Dataview 完整兼容
不做 Canvas 完整兼容
不做多人协同编辑
```

---

# 2. 核心目标

## 2.1 第一阶段目标

实现一个最小可用版本：

```text
1. 启动 Go 服务
2. 指定 Vault 目录
3. 扫描 Markdown 文件
4. 浏览器查看目录树
5. 点击 Markdown 文件后渲染内容
6. 支持 Obsidian [[双链]] 跳转
7. 支持 ![[图片.png]] 附件展示
8. 支持反向链接
9. 支持全文搜索
10. Docker 部署
```

## 2.2 运行方式

```bash
vault-reader \
  --vault /opt/obsidian-vault \
  --data /opt/vault-reader-data \
  --addr :3000
```

访问：

```text
http://server-ip:3000
```

---

# 3. 推荐技术栈

## 3.1 后端

```text
语言：Go
Web 框架：net/http + chi
Markdown 渲染：goldmark
Frontmatter：yaml.v3
文件监听：fsnotify
数据库：SQLite
全文搜索：SQLite FTS5
配置：命令行参数 + 环境变量
日志：slog
```

推荐依赖：

```bash
go get github.com/go-chi/chi/v5
go get github.com/yuin/goldmark
go get github.com/yuin/goldmark/extension
go get github.com/yuin/goldmark/parser
go get github.com/yuin/goldmark/renderer/html
go get gopkg.in/yaml.v3
go get github.com/fsnotify/fsnotify
go get github.com/mattn/go-sqlite3
```

注意：

```text
github.com/mattn/go-sqlite3 需要 CGO
如果希望纯静态编译，可后续改为 modernc.org/sqlite
```

## 3.2 前端

第一版不必上复杂前端框架，建议：

```text
HTML + CSS + 少量 Vanilla JS
```

原因：

```text
部署简单
不需要 Node 构建链
Go 可直接 embed 静态资源
运行成本低
适合只读知识库
```

后续如果 UI 复杂，再升级到 React / Vue。

---

# 4. 总体架构

```text
┌────────────────────────────┐
│        Obsidian Vault       │
│  /opt/obsidian-vault        │
│                            │
│  00_Inbox                  │
│  10_Reference              │
│  20_Debug                  │
│  30_Dashboard              │
│  90_Templates              │
└──────────────┬─────────────┘
               │ 只读扫描
               ▼
┌────────────────────────────┐
│       Vault Reader          │
│                            │
│  Scanner                   │
│  Parser                    │
│  Link Resolver             │
│  Indexer                   │
│  Search                    │
│  HTTP Server               │
└──────────────┬─────────────┘
               │
               ▼
┌────────────────────────────┐
│         Browser UI          │
│                            │
│  目录树                     │
│  Markdown 阅读              │
│  搜索                       │
│  反链                       │
│  标签                       │
└────────────────────────────┘
```

---

# 5. 项目目录结构

```text
vault-reader/
├─ cmd/
│  └─ vault-reader/
│     └─ main.go
│
├─ internal/
│  ├─ config/
│  │  └─ config.go
│  │
│  ├─ scanner/
│  │  ├─ scanner.go
│  │  └─ tree.go
│  │
│  ├─ parser/
│  │  ├─ markdown.go
│  │  ├─ frontmatter.go
│  │  ├─ wikilink.go
│  │  └─ heading.go
│  │
│  ├─ resolver/
│  │  └─ resolver.go
│  │
│  ├─ indexer/
│  │  ├─ indexer.go
│  │  ├─ schema.go
│  │  └─ repository.go
│  │
│  ├─ search/
│  │  └─ search.go
│  │
│  ├─ server/
│  │  ├─ server.go
│  │  ├─ routes.go
│  │  ├─ handlers_file.go
│  │  ├─ handlers_tree.go
│  │  ├─ handlers_search.go
│  │  ├─ handlers_assets.go
│  │  └─ middleware.go
│  │
│  └─ security/
│     └─ path.go
│
├─ web/
│  ├─ index.html
│  ├─ app.js
│  └─ style.css
│
├─ scripts/
│  ├─ build.sh
│  └─ install-service.sh
│
├─ Dockerfile
├─ docker-compose.yml
├─ go.mod
├─ go.sum
└─ README.md
```

---

# 6. 功能模块拆解

## 6.1 配置模块

### 目标

支持从命令行参数和环境变量读取配置。

### 配置项

```text
VAULT_DIR      Obsidian Vault 路径
DATA_DIR       索引数据库路径
ADDR           服务监听地址
BASE_URL       可选，反代路径
READONLY       默认 true
AUTH_ENABLED   是否启用简单认证，后续阶段
```

### 命令示例

```bash
vault-reader --vault /opt/obsidian-vault --data /opt/vault-reader-data --addr :3000
```

### 交付物

```text
internal/config/config.go
```

---

## 6.2 文件扫描模块

### 目标

递归扫描 Vault 下的文件，识别 Markdown、图片、PDF 和附件。

### 支持文件

Markdown：

```text
.md
.markdown
```

附件：

```text
.png
.jpg
.jpeg
.webp
.gif
.svg
.pdf
.txt
json
yaml
yml
zip
```

### 需要忽略

```text
.obsidian/
.trash/
.git/
node_modules/
.DS_Store
```

### 输出结构

```go
type VaultFile struct {
    Path       string
    AbsPath    string
    Name       string
    Ext        string
    IsMarkdown bool
    Size       int64
    ModTime    time.Time
}
```

### 交付物

```text
internal/scanner/scanner.go
internal/scanner/tree.go
```

---

## 6.3 Markdown 解析模块

### 目标

把 Markdown 原文解析为：

```text
frontmatter
正文
标题
heading 列表
标签
wikilinks
embeds
纯文本内容
HTML
```

### 需要支持

```md
---
title: OpenClaw 配置排查
tags:
  - openclaw
  - debug
---

# OpenClaw 配置排查

参考 [[OpenClaw]]
查看 [[OpenClaw#代理配置|代理配置]]
图片 ![[架构图.png]]
```

### 输出结构

```go
type ParsedDocument struct {
    Path        string
    Title       string
    Frontmatter map[string]any
    Content     string
    PlainText   string
    HTML        string
    Headings    []Heading
    Links       []WikiLink
    Tags        []string
}
```

### 交付物

```text
internal/parser/markdown.go
internal/parser/frontmatter.go
internal/parser/wikilink.go
internal/parser/heading.go
```

---

## 6.4 Obsidian 双链解析模块

### 目标

兼容常见 Obsidian 双链语法。

### 必须支持

```md
[[Page]]
[[Page|Alias]]
[[Page#Heading]]
[[Page#Heading|Alias]]
[[folder/Page]]
[[folder/Page|Alias]]
![[image.png]]
![[folder/image.png]]
![[Page]]
```

### 解析结构

```go
type WikiLink struct {
    Raw       string
    Target    string
    Alias     string
    Heading   string
    IsEmbed   bool
    IsAsset   bool
}
```

### 解析规则

```text
[[OpenClaw]]
target = OpenClaw

[[OpenClaw|排查文档]]
target = OpenClaw
alias = 排查文档

[[OpenClaw#代理配置]]
target = OpenClaw
heading = 代理配置

[[OpenClaw#代理配置|代理说明]]
target = OpenClaw
heading = 代理配置
alias = 代理说明

![[架构图.png]]
target = 架构图.png
isEmbed = true
isAsset = true
```

### 交付物

```text
internal/parser/wikilink.go
```

---

## 6.5 链接解析 Resolver

### 目标

把 `[[OpenClaw]]` 解析到真实文件路径。

### 匹配优先级

```text
1. 精确相对路径匹配
   [[20_Debug/OpenClaw/代理排查]]

2. 去除 .md 后精确路径匹配
   [[20_Debug/OpenClaw/代理排查]]

3. 文件名匹配
   [[OpenClaw]] → **/OpenClaw.md

4. frontmatter title 匹配
   title: OpenClaw

5. 一级标题匹配
   # OpenClaw

6. 多个候选时返回冲突列表
```

### 结果结构

```go
type ResolveResult struct {
    Found      bool
    TargetPath string
    Candidates []string
    IsAmbiguous bool
}
```

### 页面表现

如果找到：

```html
<a href="/note?path=20_Debug/OpenClaw.md">OpenClaw</a>
```

如果未找到：

```html
<span class="broken-link">OpenClaw</span>
```

如果多个候选：

```html
<span class="ambiguous-link">OpenClaw</span>
```

点击后显示候选列表。

### 交付物

```text
internal/resolver/resolver.go
```

---

## 6.6 索引模块

### 目标

把 Vault 内容索引进 SQLite，便于查询、反链、搜索。

### 数据库位置

```text
/data/vault-reader.db
```

### 表设计

#### files

```sql
CREATE TABLE IF NOT EXISTS files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  ext TEXT NOT NULL,
  size INTEGER NOT NULL,
  mtime INTEGER NOT NULL,
  content TEXT,
  html TEXT,
  frontmatter_json TEXT,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
```

#### links

```sql
CREATE TABLE IF NOT EXISTS links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_path TEXT NOT NULL,
  raw TEXT NOT NULL,
  target TEXT NOT NULL,
  target_path TEXT,
  alias TEXT,
  heading TEXT,
  is_embed INTEGER NOT NULL DEFAULT 0,
  is_asset INTEGER NOT NULL DEFAULT 0,
  resolved INTEGER NOT NULL DEFAULT 0
);
```

#### tags

```sql
CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  tag TEXT NOT NULL
);
```

#### headings

```sql
CREATE TABLE IF NOT EXISTS headings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  level INTEGER NOT NULL,
  text TEXT NOT NULL,
  slug TEXT NOT NULL
);
```

#### search index

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS file_fts USING fts5(
  title,
  path,
  content
);
```

### 交付物

```text
internal/indexer/schema.go
internal/indexer/indexer.go
internal/indexer/repository.go
```

---

## 6.7 反链模块

### 目标

打开某篇笔记时，显示所有引用它的笔记。

### 查询逻辑

```sql
SELECT from_path, raw, alias
FROM links
WHERE target_path = ?
ORDER BY from_path;
```

### UI 展示

```text
Backlinks
├─ 20_Debug/OpenClaw/代理排查.md
│  └─ [[OpenClaw]]
├─ 10_Reference/官方文档.md
│  └─ [[OpenClaw|OpenClaw 文档]]
```

### 交付物

```text
GET /api/backlinks?path=xxx
```

---

## 6.8 搜索模块

### 目标

支持标题、路径、正文搜索。

### API

```text
GET /api/search?q=openclaw
```

### 返回结构

```json
{
  "items": [
    {
      "path": "20_Debug/OpenClaw/代理排查.md",
      "title": "OpenClaw 代理排查",
      "snippet": "...HTTP_PROXY..."
    }
  ]
}
```

### 第一版搜索策略

使用 SQLite FTS5：

```sql
SELECT path, title, snippet(file_fts, 2, '<mark>', '</mark>', '...', 10)
FROM file_fts
WHERE file_fts MATCH ?
LIMIT 50;
```

### 中文搜索说明

SQLite FTS5 对中文分词能力有限，第一版可以接受：

```text
精确中文词搜索
英文关键词搜索
路径搜索
标题搜索
```

后续优化：

```text
接入 Meilisearch
接入 Bleve
接入 Tantivy
自建中文分词
```

---

## 6.9 附件访问模块

### 目标

支持安全访问 Vault 中的图片、PDF、文本附件。

### API

```text
GET /assets?path=xxx
```

### 安全要求

必须防止路径穿越：

```text
禁止 ../../etc/passwd
禁止访问 Vault 外部文件
只允许访问 VAULT_DIR 内文件
```

### 支持展示

图片：

```html
<img src="/assets?path=attachments/a.png">
```

PDF：

```html
<iframe src="/assets?path=docs/a.pdf"></iframe>
```

其他附件：

```html
<a href="/assets?path=file.zip">下载附件</a>
```

### 交付物

```text
internal/server/handlers_assets.go
internal/security/path.go
```

---

## 6.10 文件监听模块

### 目标

Vault 文件变化后自动更新索引。

### 使用

```text
fsnotify
```

### 监听事件

```text
新增 .md
修改 .md
删除 .md
重命名 .md
新增附件
删除附件
```

### 更新策略

第一版：

```text
检测到变化后 debounce 2 秒
重新扫描全库
重新构建索引
```

第二版：

```text
增量更新单文件
```

### 交付物

```text
internal/indexer/watcher.go
```

---

# 7. Web UI 设计

## 7.1 页面布局

```text
┌──────────────────────────────────────────────────────────────┐
│ Vault Reader                         Search: [             ] │
├────────────────────┬──────────────────────────────┬──────────┤
│ 文件树              │ Markdown 内容                 │ 右侧栏    │
│                    │                              │          │
│ 00_Inbox           │ # OpenClaw 配置排查            │ TOC      │
│ 10_Reference       │                              │ Tags     │
│ 20_Debug           │ [[Codex]]                     │ Backlinks│
│ 30_Dashboard       │                              │          │
└────────────────────┴──────────────────────────────┴──────────┘
```

## 7.2 必要页面

```text
首页
笔记详情页
搜索结果页
标签页
404 / 未找到页
```

## 7.3 前端 API

```text
GET /api/tree
GET /api/note?path=xxx
GET /api/search?q=xxx
GET /api/backlinks?path=xxx
GET /api/tags
GET /api/tag?name=xxx
GET /assets?path=xxx
```

## 7.4 UI 第一版要求

```text
支持暗色模式
支持代码块样式
支持表格
支持任务列表
支持移动端基本阅读
搜索结果高亮
破损链接显示为红色
未解析链接显示提示
```

---

# 8. API 设计

## 8.1 获取目录树

```http
GET /api/tree
```

返回：

```json
{
  "items": [
    {
      "name": "20_Debug",
      "path": "20_Debug",
      "type": "dir",
      "children": [
        {
          "name": "OpenClaw.md",
          "path": "20_Debug/OpenClaw.md",
          "type": "file"
        }
      ]
    }
  ]
}
```

---

## 8.2 获取笔记

```http
GET /api/note?path=20_Debug/OpenClaw.md
```

返回：

```json
{
  "path": "20_Debug/OpenClaw.md",
  "title": "OpenClaw",
  "html": "<h1>OpenClaw</h1>",
  "frontmatter": {},
  "links": [],
  "tags": ["openclaw"],
  "headings": [],
  "backlinks": []
}
```

---

## 8.3 搜索

```http
GET /api/search?q=proxy
```

返回：

```json
{
  "items": [
    {
      "path": "20_Debug/OpenClaw/代理排查.md",
      "title": "OpenClaw 代理排查",
      "snippet": "...<mark>proxy</mark>..."
    }
  ]
}
```

---

## 8.4 获取反链

```http
GET /api/backlinks?path=20_Debug/OpenClaw.md
```

返回：

```json
{
  "items": [
    {
      "fromPath": "20_Debug/Codex.md",
      "title": "Codex",
      "raw": "[[OpenClaw]]"
    }
  ]
}
```

---

## 8.5 获取标签

```http
GET /api/tags
```

返回：

```json
{
  "items": [
    {
      "tag": "openclaw",
      "count": 12
    }
  ]
}
```

---

# 9. 安全要求

## 9.1 必须只读

```text
服务不能修改 Vault 原文件
Docker 挂载必须 :ro
后端不提供写文件 API
```

## 9.2 路径安全

必须实现：

```text
clean path
realpath 校验
禁止 ..
禁止绝对路径参数
禁止访问 Vault 外路径
```

示例：

```text
/assets?path=../../etc/passwd
```

必须返回：

```text
403 Forbidden
```

## 9.3 隐私与访问控制

第一版可以用 Nginx Basic Auth。

后续服务内置：

```text
Basic Auth
Bearer Token
IP allowlist
```

---

# 10. Docker 部署

## 10.1 Dockerfile

目标：

```text
多阶段构建
最终镜像尽量小
只包含 vault-reader 二进制
```

## 10.2 docker-compose.yml

```yaml
services:
  vault-reader:
    image: vault-reader:latest
    container_name: vault-reader
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      VAULT_DIR: /vault
      DATA_DIR: /data
      ADDR: ":3000"
    volumes:
      - /opt/obsidian-vault:/vault:ro
      - /opt/vault-reader-data:/data
```

---

# 11. Nginx 反向代理

```nginx
server {
    listen 80;
    server_name kb.example.com;

    auth_basic "Knowledge Base";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

生成密码：

```bash
sudo apt install apache2-utils -y
sudo htpasswd -c /etc/nginx/.htpasswd qsc
sudo nginx -t
sudo systemctl reload nginx
```

---

# 12. 开发里程碑

## Milestone 0：项目初始化

### 目标

创建 Go 项目骨架。

### 任务

```text
1. 创建 go.mod
2. 创建 cmd/vault-reader/main.go
3. 创建 internal/config
4. 创建 web/index.html
5. 启动 HTTP 服务
6. 返回首页
```

### 验收标准

```bash
go run ./cmd/vault-reader --vault ./test-vault
```

访问：

```text
http://127.0.0.1:3000
```

能看到首页。

---

## Milestone 1：Vault 扫描与目录树

### 目标

扫描 Vault 并展示目录树。

### 任务

```text
1. 递归扫描 Vault
2. 忽略 .obsidian、.git、node_modules
3. 识别 md 文件
4. 实现 /api/tree
5. 前端展示目录树
```

### 验收标准

```text
左侧能看到：
00_Inbox
10_Reference
20_Debug
30_Dashboard
90_Templates
```

---

## Milestone 2：Markdown 渲染

### 目标

点击文件后显示 Markdown 渲染内容。

### 任务

```text
1. 读取 md 文件
2. 解析 frontmatter
3. 使用 goldmark 渲染 HTML
4. 支持 GFM 表格、代码块、任务列表
5. 实现 /api/note
6. 前端点击文件显示内容
```

### 验收标准

```text
# 标题正常
代码块正常
表格正常
任务列表正常
frontmatter 不显示在正文中
```

---

## Milestone 3：Obsidian 双链

### 目标

支持 `[[双链]]` 点击跳转。

### 任务

```text
1. 编写 wikilink parser
2. 支持 [[Page]]
3. 支持 [[Page|Alias]]
4. 支持 [[Page#Heading]]
5. 支持 [[Page#Heading|Alias]]
6. 建立文件名到路径的索引
7. 渲染为可点击链接
8. 处理未找到链接
9. 处理多候选链接
```

### 验收标准

```md
[[OpenClaw]]
[[OpenClaw|OpenClaw 排查]]
[[OpenClaw#代理配置]]
```

都能正常显示并跳转。

---

## Milestone 4：附件嵌入

### 目标

支持图片和 PDF 附件。

### 任务

```text
1. 解析 ![[image.png]]
2. 解析 ![[folder/image.png]]
3. 判断附件类型
4. 图片渲染 img
5. PDF 渲染 iframe
6. 其他附件渲染下载链接
7. 实现 /assets 安全访问
```

### 验收标准

```md
![[架构图.png]]
![[docs/spec.pdf]]
```

能正常显示。

---

## Milestone 5：索引与反链

### 目标

建立 SQLite 索引并显示反链。

### 任务

```text
1. 初始化 SQLite schema
2. 扫描所有 Markdown
3. 保存 files
4. 保存 links
5. 保存 tags
6. 保存 headings
7. 实现 target_path 解析
8. 实现 /api/backlinks
9. 右侧栏展示反链
```

### 验收标准

打开 `OpenClaw.md` 时，右侧能显示哪些笔记引用了它。

---

## Milestone 6：全文搜索

### 目标

实现标题、路径、正文搜索。

### 任务

```text
1. 初始化 FTS5 表
2. 写入标题、路径、正文
3. 实现 /api/search
4. 搜索结果展示标题、路径、片段
5. 点击结果打开笔记
```

### 验收标准

搜索：

```text
proxy
OpenClaw
Docker
Feishu
```

能返回相关文档。

---

## Milestone 7：文件监听与索引更新

### 目标

Vault 文件变化后自动更新。

### 任务

```text
1. 使用 fsnotify 监听 Vault
2. debounce 防抖
3. 文件变化后重建索引
4. 前端刷新后看到最新内容
```

### 验收标准

新增一个 Markdown 文件后，不重启服务也能在目录树中看到。

---

## Milestone 8：部署与文档

### 目标

可在 Linux 服务器稳定部署。

### 任务

```text
1. 编写 Dockerfile
2. 编写 docker-compose.yml
3. 编写 README
4. 提供 systemd service 示例
5. 提供 Nginx 反代示例
6. 提供备份说明
```

### 验收标准

可以通过：

```bash
docker compose up -d
```

启动服务并访问知识库。

---

# 13. AI Agent 执行任务 Prompt

可以直接交给 AI Agent：

```text
你要开发一个 Go 语言实现的只读 Obsidian Vault Web Reader。

项目目标：
- 运行在 Linux 服务器
- 读取已有 Obsidian Vault
- 不修改 Vault 原文件
- 通过浏览器查看 Markdown 知识库
- 高兼容 Obsidian 双链
- 支持反链、搜索、标签、附件预览
- 单进程运行，资源占用低

技术要求：
- Go
- net/http + chi
- goldmark 渲染 Markdown
- yaml.v3 解析 frontmatter
- SQLite 保存索引
- SQLite FTS5 做全文搜索
- fsnotify 监听文件变化
- 前端使用 HTML + CSS + Vanilla JS，不引入 Node 构建链
- 支持 Docker 部署

第一阶段先完成：
1. 项目骨架
2. 配置读取
3. Vault 目录扫描
4. /api/tree
5. /api/note
6. Markdown 渲染
7. 前端目录树和正文查看

注意：
- Vault 必须只读
- 必须防止路径穿越
- 忽略 .obsidian、.git、node_modules
- 支持中文文件名和空格路径
- README 写清楚如何运行
```

---

# 14. 验收用测试 Vault

建议准备一个 `test-vault`：

```text
test-vault/
├─ 00_Inbox/
│  └─ 临时记录.md
├─ 10_Reference/
│  └─ 官方文档.md
├─ 20_Debug/
│  ├─ OpenClaw.md
│  ├─ Codex.md
│  └─ Docker.md
├─ 30_Dashboard/
│  └─ 常用系统入口.md
├─ attachments/
│  └─ 架构图.png
└─ 90_Templates/
   └─ Web Clip 模板.md
```

测试内容：

```md
---
title: OpenClaw 配置排查
tags:
  - openclaw
  - debug
---

# OpenClaw 配置排查

参考 [[Codex]]
查看 [[Docker|Docker 排查]]

## 代理配置

这里记录 HTTP_PROXY 和 HTTPS_PROXY。

图片：

![[attachments/架构图.png]]
```

---

# 15. 风险点与规避

## 15.1 中文搜索不理想

风险：

```text
SQLite FTS5 中文分词弱
```

规避：

```text
第一版接受精确搜索
后续接 Meilisearch / Bleve
```

## 15.2 Obsidian 链接解析复杂

风险：

```text
同名文件
别名
标题链接
路径中有空格和中文
```

规避：

```text
先支持常用语法
多候选时不强行猜测
页面提示候选列表
```

## 15.3 附件路径安全

风险：

```text
/assets?path=../../etc/passwd
```

规避：

```text
realpath 校验
限制在 VAULT_DIR 内
Docker 只读挂载
```

## 15.4 Vault 很大时首次索引慢

规避：

```text
显示索引状态
后台构建
按 mtime 增量更新
```

---

# 16. 版本规划

## v0.1

```text
目录树
Markdown 渲染
基础页面
Docker 部署
```

## v0.2

```text
Obsidian 双链
附件预览
路径安全
```

## v0.3

```text
SQLite 索引
反链
标签
```

## v0.4

```text
全文搜索
搜索高亮
文件监听
```

## v0.5

```text
UI 优化
暗色模式
Nginx 示例
systemd 示例
```

## v1.0

```text
稳定只读知识库
支持大多数常见 Obsidian Vault
可长期部署
```

---

# 17. 最终交付清单

```text
1. Go 源码
2. 单二进制构建脚本
3. Dockerfile
4. docker-compose.yml
5. README.md
6. Nginx 配置示例
7. systemd service 示例
8. test-vault 示例
9. API 文档
10. 安全说明
```

---

# 18. 推荐实施顺序

最稳妥的顺序是：

```text
先跑通：目录树 + Markdown 渲染
再增强：双链 + 附件
再沉淀：索引 + 反链
再体验：搜索 + 标签 + UI
最后部署：Docker + Nginx + Basic Auth
```

一句话目标：

> **做一个轻量、只读、单进程、低成本、兼容 Obsidian 双链的服务器端 Markdown 知识库浏览器。**
