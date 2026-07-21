# Eino Career Agent 职引

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Eino](https://img.shields.io/badge/Eino-ADK-00BFFF?style=flat)](https://github.com/cloudwego/eino)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**职引** — 基于 ByteDance Eino ADK 构建的 AI 职业规划 Agent 应用。单一聊天入口，Agent 自动识别用户意图并路由到对应处理流程，无需前端模式切换。

An AI career planning agent powered by ByteDance Eino ADK. A single chat entry point with automatic intent routing — the agent decides whether to answer questions, generate reports, or parse resumes.

---

## Features ✨ 功能特性

- 🤖 **Agent Auto-Routing** — 基于 Eino ADK `ChatModelAgent` 的 ReAct 循环，自动识别用户意图（问答/生成报告/解析简历/查询报告），无需前端模式切换
- 📊 **12-Section Parallel Report** — 职业规划报告 12 章节并行生成，goroutine + semaphore 控制并发，总耗时接近单章节耗时
- 📄 **Resume Parsing** — 支持 PDF/DOCX 简历文件上传解析，Agent 自动调用 `parse_resume_file` Tool 提取文本内容
- 🌊 **SSE Streaming** — 全链路流式输出，前端实时展示 Agent 推理过程与 Tool 调用结果
- 🔌 **OpenAI-Compatible** — 通过 `OPENAI_BASE_URL` + `OPENAI_API_KEY` + `OPENAI_MODEL` 环境变量接入任意 OpenAI 兼容 LLM（ModelScope / DeepSeek / 通义千问 / OpenAI 等）
- 📦 **Zero-Dependency Deploy** — 纯 Go + SQLite，无需 Redis/MongoDB/Python/Java，单二进制 + 前端 SPA 一体化部署

---

## Tech Stack 🛠️ 技术栈

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend 语言** | Go | 1.22+ |
| **AI 框架** | Eino ADK (cloudwego/eino) | v0.9.12 |
| **Web 框架** | Gin | v1.12+ |
| **数据库** | SQLite (mattn/go-sqlite3 + GORM) | - |
| **认证** | JWT (golang-jwt/jwt/v5) | v5 |
| **前端框架** | Vue 3 + Vite | 3.4+ / 5+ |
| **UI 库** | Element Plus | - |
| **状态管理** | Pinia | - |
| **HTTP 客户端** | Axios | - |

---

## Architecture 🏗️ 架构

```
User Chat Input
      │
      ▼
┌─────────────────────────────────────┐
│        Eino ChatModelAgent          │
│     (ReAct Loop, MaxIterations=10)  │
│                                     │
│   ┌─────────┐    ┌──────────────┐  │
│   │ ChatModel│◄──►│ Tool Router  │  │
│   └─────────┘    └──────┬───────┘  │
│                         │           │
│        ┌────────────────┼───────┐   │
│        ▼                ▼       ▼   │
│  ┌──────────┐  ┌──────────┐ ┌─────┐│
│  │generate_  │  │parse_    │ │query││
│  │career_    │  │resume_   │ │career││
│  │report     │  │file      │ │report││
│  └─────┬─────┘  └────┬─────┘ └──┬──┘│
│        │              │          │   │
│        ▼              ▼          ▼   │
│  12-Section     PDF/DOCX     Report  │
│  Parallel       Parser       Query   │
│  Generation                         │
└─────────────────────────────────────┘
      │
      ▼
   SSE Stream → Frontend
```

### Project Structure 项目目录

```
eino-carrer-agent/
├── LICENSE                           # MIT License
├── docs/                             # 文档
├── server/                           # Go 后端
│   ├── cmd/server/main.go            # 程序入口（依赖注入+路由+SPA托管）
│   ├── configs/config.yaml           # 配置文件
│   └── internal/
│       ├── config/config.go          # 配置加载（viper+环境变量）
│       ├── domain/domain.go          # 5个 GORM 模型
│       ├── repository/               # 5个 Repository（User/Conv/Msg/Report/File）
│       ├── handler/                  # HTTP Handler（Auth/Chat/Upload/Report/Conv/Middleware）
│       ├── service/                  # 业务逻辑（Auth/Chat/Upload）
│       ├── agent/                    # Eino Agent 核心
│       │   ├── agent.go              # CareerAgent（ChatModelAgent + 3 Tools + StreamChat）
│       │   ├── prompts/              # SystemPrompt + 12章节Prompt
│       │   ├── tools/                # 3个 Agent Tools
│       │   ├── graph/                # 报告并行生成（goroutine+semaphore）
│       │   └── callback/             # Callback 日志
│       ├── parser/                   # PDF/DOCX 解析器
│       └── pkg/                      # 内部工具（jwt/sse/response/database/logger）
├── web/                              # Vue3 前端
│   ├── src/
│   │   ├── api/                      # Axios API 封装
│   │   ├── components/               # 5个组件（ChatWindow/MessageBubble/ReportView/FileUpload/LoginDialog）
│   │   ├── stores/                   # Pinia Stores（user/chat）
│   │   └── utils/                    # SSE客户端 + Axios封装 + Token管理
│   ├── vite.config.js
│   └── package.json
```

---

## Quick Start 🚀 快速开始

### Prerequisites 前置要求

- Go 1.22+
- Node.js 18+（仅前端开发需要）
- 一个 OpenAI 兼容的 LLM API Key

### 1. Clone & Setup 克隆项目

```bash
git clone https://github.com/hautmz/eino-carrer-agent.git
cd eino-carrer-agent
```

### 2. Configure LLM 配置 LLM 环境变量

```bash
# 必须设置，用于接入 OpenAI 兼容 LLM
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"  # 示例：通义千问
export OPENAI_MODEL="qwen-plus"                                             # 示例模型名
```

Windows PowerShell:
```powershell
$env:OPENAI_API_KEY = "your-api-key"
$env:OPENAI_BASE_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
$env:OPENAI_MODEL = "qwen-plus"
```

### 3. Run 运行

**一体化运行（后端托管前端 SPA）:**

```bash
# 先构建前端
cd web && npm install && npm run build && cd ..

# 启动后端（自动托管 web/dist）
cd server && go run ./cmd/server
```

访问 `http://localhost:8081` 即可使用。

**前后端分离开发:**

```bash
# Terminal 1: 后端
cd server && go run ./cmd/server

# Terminal 2: 前端（Vite dev server，热更新）
cd web && npm run dev
```

---

## Configuration ⚙️ 配置

### config.yaml

配置文件位于 `server/configs/config.yaml`，关键配置项：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `server.port` | `8081` | 服务监听端口 |
| `server.mode` | `debug` | 运行模式: debug / release / test |
| `database.path` | `./data/eino_career.db` | SQLite 数据库文件路径 |
| `jwt.secret` | `eino-career-agent-...` | JWT 签名密钥，**生产环境必须修改** |
| `jwt.expiration` | `72h` | Token 过期时间 |
| `upload.max_size` | `10485760` | 文件最大大小（字节），默认 10MB |
| `upload.allowed_types` | `["pdf","docx"]` | 允许的文件类型 |
| `upload.storage_path` | `./data/uploads` | 文件存储路径 |
| `agent.report_timeout` | `360` | 报告生成总超时（秒） |
| `agent.section_timeout` | `120` | 单章节生成超时（秒） |
| `agent.max_concurrent_sections` | `4` | 报告章节最大并行数 |
| `agent.max_history_messages` | `50` | 单次对话加载的最大历史消息数 |
| `sse.heartbeat_interval` | `15` | SSE 心跳间隔（秒） |

### Environment Variables 环境变量

| 环境变量 | 必填 | 说明 |
|----------|------|------|
| `OPENAI_API_KEY` | ✅ | LLM API Key |
| `OPENAI_BASE_URL` | ✅ | OpenAI 兼容 API Base URL |
| `OPENAI_MODEL` | ✅ | 模型名称 |

支持任意 OpenAI 兼容 API，例如：

| Provider | Base URL 示例 |
|----------|--------------|
| 阿里云通义千问 | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| DeepSeek | `https://api.deepseek.com/v1` |
| ModelScope | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| OpenAI | `https://api.openai.com/v1` |
| 本地 Ollama | `http://localhost:11434/v1` |

---

## API Reference 📡 API 接口

### Auth 认证

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录，返回 JWT Token |

### Chat 聊天

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/chat/stream` | SSE 流式聊天，Agent 自动意图路由 |

请求体：
```json
{
  "conversation_id": "",
  "message": "帮我生成职业规划报告",
  "file_id": null
}
```

SSE 事件类型：`message` | `tool_call` | `error` | `done`

### Upload 文件上传

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/upload` | 上传文件（PDF/DOCX） |
| GET | `/api/upload/:id` | 获取文件信息 |

### Report 报告

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/report/list` | 获取当前用户的报告列表 |
| GET | `/api/report/:id` | 获取报告详情（含12章节内容） |

### Conversation 对话

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/conversation/list` | 获取对话列表 |
| GET | `/api/conversation/:id` | 获取对话详情（含消息） |
| DELETE | `/api/conversation/:id` | 删除对话 |

> 所有 `/api/` 开头的接口（除 Auth 外）均需在 Header 中携带 `Authorization: Bearer <token>`。

---

## Agent Tools 🔧 Agent 工具

Agent 基于 Eino ADK `ChatModelAgent` 的 ReAct 循环，自动判断是否需要调用工具：

| Tool Name | 触发场景 | 功能 |
|-----------|---------|------|
| `generate_career_report` | 用户要求生成职业规划报告 | 并行生成 12 章节报告，写入数据库，返回报告 ID |
| `parse_resume_file` | 用户上传了简历文件 | 解析 PDF/DOCX 文件内容，提取文本供后续分析 |
| `query_career_report` | 用户查询已有报告 | 按报告 ID 查详情，或列出用户所有报告 |

---

## Report Sections 📊 报告章节

职业规划报告包含 12 个章节，并行生成后合并：

| # | 章节标识 | 中文标题 |
|---|---------|---------|
| 1 | `professional_index` | 专业指数评分 |
| 2 | `myself_report` | 个人信息提取 |
| 3 | `achievement_superiority` | 成就与优势 |
| 4 | `career_experience` | 职业经历与成长路径 |
| 5 | `motivation_values` | 动机与价值观评估 |
| 6 | `skill_heatmap` | 技能热力图与胜任力模型 |
| 7 | `interest_assessment` | 职业兴趣评估 |
| 8 | `career_recommendations` | 基于兴趣的职业推荐 |
| 9 | `industry_analysis` | 行业分析 |
| 10 | `goal_setting` | 目标设定 |
| 11 | `action_plan` | 行动计划 |
| 12 | `summary_outlook` | 总结与展望 |

---

## Development 👩‍💻 开发

### Backend 后端

```bash
cd server

# 运行
go run ./cmd/server

# 构建
go build -o bin/server ./cmd/server

# 测试
go test ./...

# Lint
golangci-lint run
```

### Frontend 前端

```bash
cd web

# 安装依赖
npm install

# 开发模式（热更新，默认 localhost:5173）
npm run dev

# 构建生产版本
npm run build

# Lint
npm run lint
```

### Build All 全量构建

```bash
cd web && npm run build && cd ../server && go build -o bin/server ./cmd/server
```

---

## Roadmap 🗺️ 规划

- [ ] 报告 PDF 导出功能
- [ ] 多轮对话记忆优化（长期记忆 / 摘要压缩）
- [ ] 前端报告可视化渲染（雷达图/热力图）
- [ ] 多模型热切换（UI 中选择模型）
- [ ] Docker 容器化部署
- [ ] 多语言支持（i18n）

---

## Contributing 🤝 贡献

欢迎贡献！请提交 Pull Request。

- Go 代码遵循标准 `gofmt` + `golangci-lint` 风格
- 前端代码遵循 ESLint 配置
- 提交信息建议使用 `feat:` / `fix:` / `docs:` / `refactor:` 前缀

---

## License 📄 许可证

[MIT License](LICENSE) © 2025 hautmz

---

## Acknowledgements 🙏 致谢

- [cloudwego/eino](https://github.com/cloudwego/eino) — ByteDance 开源的 Go AI 应用开发框架
- [Gin](https://github.com/gin-gonic/gin) — Go HTTP Web 框架
- [Vue.js](https://vuejs.org/) — 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) — Vue 3 组件库
