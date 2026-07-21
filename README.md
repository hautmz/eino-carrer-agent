# Eino Career Agent 职引

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Eino](https://img.shields.io/badge/Eino-ADK-00BFFF?style=flat)](https://github.com/cloudwego/eino)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**职引** — 基于 ByteDance Eino ADK 构建的 AI 职业规划 Agent 应用。单一聊天入口，Agent 自动识别用户意图并路由到对应处理流程，无需前端模式切换。

**ZhiYin** — An AI career planning agent powered by ByteDance Eino ADK. A single chat entry point with automatic intent routing — the agent decides whether to answer questions, generate reports, or parse resumes, with no front-end mode switching required.

---

## Features ✨ 功能特性

- 🤖 **Agent Auto-Routing** — 基于 Eino ADK `ChatModelAgent` 的 ReAct 循环，自动识别用户意图（问答/生成报告/解析简历/查询报告），无需前端模式切换
  <br>Built on Eino ADK `ChatModelAgent` with a ReAct loop that automatically detects user intent (Q&A / report generation / resume parsing / report query) — no front-end mode switching needed.

- 📊 **12-Section Parallel Report** — 职业规划报告 12 章节并行生成，goroutine + semaphore 控制并发，总耗时接近单章节耗时
  <br>Generates 12 report sections in parallel using goroutines with semaphore-based concurrency control, keeping total time close to a single section.

- 📄 **Resume Parsing** — 支持 PDF/DOCX 简历文件上传解析，Agent 自动调用 `parse_resume_file` Tool 提取文本内容
  <br>Supports PDF/DOCX resume uploads. The agent automatically invokes the `parse_resume_file` tool to extract text content for career analysis.

- 🌊 **SSE Streaming** — 全链路流式输出，前端实时展示 Agent 推理过程与 Tool 调用结果
  <br>End-to-end Server-Sent Events streaming — the front-end displays the agent's reasoning and tool call results in real time.

- 🔌 **OpenAI-Compatible** — 通过 `OPENAI_BASE_URL` + `OPENAI_API_KEY` + `OPENAI_MODEL` 环境变量接入任意 OpenAI 兼容 LLM（ModelScope / DeepSeek / 通义千问 / OpenAI 等）
  <br>Connect to any OpenAI-compatible LLM via environment variables — works with ModelScope, DeepSeek, Qwen, OpenAI, Ollama, and more.

- 📦 **Zero-Dependency Deploy** — 纯 Go + SQLite，无需 Redis/MongoDB/Python/Java，单二进制 + 前端 SPA 一体化部署
  <br>Pure Go + SQLite with zero external dependencies (no Redis, MongoDB, Python, or Java). Single binary deployment with built-in SPA hosting.

---

## Tech Stack 🛠️ 技术栈

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend** | Go | 1.22+ |
| **AI Framework** | Eino ADK (cloudwego/eino) | v0.9.12 |
| **Web Framework** | Gin | v1.12+ |
| **Database** | SQLite (mattn/go-sqlite3 + GORM) | - |
| **Auth** | JWT (golang-jwt/jwt/v5) | v5 |
| **Frontend** | Vue 3 + Vite | 3.4+ / 5+ |
| **UI Library** | Element Plus | - |
| **State Management** | Pinia | - |
| **HTTP Client** | Axios | - |

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
├── docs/                             # Documentation
├── server/                           # Go backend
│   ├── cmd/server/main.go            # Entry point (DI + routing + SPA hosting)
│   ├── configs/config.yaml           # Configuration file
│   └── internal/
│       ├── config/config.go          # Config loader (viper + env vars)
│       ├── domain/domain.go          # 5 GORM models
│       ├── repository/               # 5 repositories (User/Conv/Msg/Report/File)
│       ├── handler/                  # HTTP handlers (Auth/Chat/Upload/Report/Conv/Middleware)
│       ├── service/                  # Business logic (Auth/Chat/Upload)
│       ├── agent/                    # Eino Agent core
│       │   ├── agent.go              # CareerAgent (ChatModelAgent + 3 Tools + StreamChat)
│       │   ├── prompts/              # SystemPrompt + 12-section prompts
│       │   ├── tools/                # 3 Agent Tools
│       │   ├── graph/                # Parallel report generation (goroutine + semaphore)
│       │   └── callback/             # Callback logging
│       ├── parser/                   # PDF/DOCX parsers
│       └── pkg/                      # Internal utils (jwt/sse/response/database/logger)
├── web/                              # Vue3 frontend
│   ├── src/
│   │   ├── api/                      # Axios API wrappers
│   │   ├── components/               # 5 components (ChatWindow/MessageBubble/ReportView/FileUpload/LoginDialog)
│   │   ├── stores/                   # Pinia stores (user/chat)
│   │   └── utils/                    # SSE client + Axios wrapper + Token manager
│   ├── vite.config.js
│   └── package.json
```

---

## Quick Start 🚀 快速开始

### Prerequisites 前置要求

- Go 1.22+
- Node.js 18+ (only needed for front-end development)
- An OpenAI-compatible LLM API Key

### 1. Clone & Setup 克隆项目

```bash
git clone https://github.com/hautmz/eino-carrer-agent.git
cd eino-carrer-agent
```

### 2. Configure LLM 配置 LLM 环境变量

```bash
# Required — used to connect to any OpenAI-compatible LLM
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"  # Example: Qwen
export OPENAI_MODEL="qwen-plus"                                             # Example model
```

Windows PowerShell:
```powershell
$env:OPENAI_API_KEY = "your-api-key"
$env:OPENAI_BASE_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
$env:OPENAI_MODEL = "qwen-plus"
```

### 3. Run 运行

**All-in-one (backend serves front-end SPA):**

```bash
# Build the frontend first
cd web && npm install && npm run build && cd ..

# Start the backend (auto-serves web/dist)
cd server && go run ./cmd/server
```

Visit `http://localhost:8081` to start using the app.

**Separate development (hot-reload):**

```bash
# Terminal 1: Backend
cd server && go run ./cmd/server

# Terminal 2: Frontend (Vite dev server with HMR)
cd web && npm run dev
```

---

## Configuration ⚙️ 配置

### config.yaml

Located at `server/configs/config.yaml`:

| Key | Default | Description |
|-----|---------|-------------|
| `server.port` | `8081` | HTTP listen port |
| `server.mode` | `debug` | Run mode: debug / release / test |
| `database.path` | `./data/eino_career.db` | SQLite database file path |
| `jwt.secret` | `eino-career-agent-...` | JWT signing secret — **must change in production** |
| `jwt.expiration` | `72h` | Token expiration duration |
| `upload.max_size` | `10485760` | Max file size in bytes (default 10 MB) |
| `upload.allowed_types` | `["pdf","docx"]` | Allowed file extensions |
| `upload.storage_path` | `./data/uploads` | File storage directory |
| `agent.report_timeout` | `360` | Report generation timeout (seconds) |
| `agent.section_timeout` | `120` | Single section timeout (seconds) |
| `agent.max_concurrent_sections` | `4` | Max parallel section generators |
| `agent.max_history_messages` | `50` | Max history messages per conversation |
| `sse.heartbeat_interval` | `15` | SSE heartbeat interval (seconds) |

### Environment Variables 环境变量

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENAI_API_KEY` | ✅ | LLM API Key |
| `OPENAI_BASE_URL` | ✅ | OpenAI-compatible API base URL |
| `OPENAI_MODEL` | ✅ | Model name |

Compatible providers:

| Provider | Example Base URL |
|----------|-----------------|
| Alibaba Qwen / 通义千问 | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| DeepSeek | `https://api.deepseek.com/v1` |
| ModelScope | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| OpenAI | `https://api.openai.com/v1` |
| Local Ollama | `http://localhost:11434/v1` |

---

## API Reference 📡 API 接口

### Auth 认证

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and receive JWT token |

### Chat 聊天

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/chat/stream` | SSE streaming chat with agent auto-routing |

Request body:
```json
{
  "conversation_id": "",
  "message": "帮我生成职业规划报告",
  "file_id": null
}
```

SSE event types: `message` | `tool_call` | `error` | `done`

### Upload 文件上传

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/upload` | Upload a file (PDF/DOCX) |
| GET | `/api/upload/:id` | Get file info by ID |

### Report 报告

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/report/list` | List current user's reports |
| GET | `/api/report/:id` | Get report details (all 12 sections) |

### Conversation 对话

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/conversation/list` | List conversations |
| GET | `/api/conversation/:id` | Get conversation with messages |
| DELETE | `/api/conversation/:id` | Delete a conversation |

> All `/api/` endpoints (except Auth) require `Authorization: Bearer <token>` header.

---

## Agent Tools 🔧 Agent 工具

The agent uses a ReAct loop via Eino ADK `ChatModelAgent` to decide when to invoke tools:

| Tool Name | Trigger | Function |
|-----------|---------|----------|
| `generate_career_report` | User asks to generate a career report | Generates 12 sections in parallel, persists to DB, returns report ID |
| `parse_resume_file` | User uploads a resume file | Parses PDF/DOCX content for downstream career analysis |
| `query_career_report` | User asks about existing reports | Queries by report ID for details, or lists all user reports |

---

## Report Sections 📊 报告章节

The career planning report contains 12 sections, generated in parallel and merged:

| # | Section ID | Title |
|---|-----------|-------|
| 1 | `professional_index` | Professional Index Score / 专业指数评分 |
| 2 | `myself_report` | Personal Profile / 个人信息提取 |
| 3 | `achievement_superiority` | Achievements & Strengths / 成就与优势 |
| 4 | `career_experience` | Career Experience & Growth Path / 职业经历与成长路径 |
| 5 | `motivation_values` | Motivation & Values / 动机与价值观评估 |
| 6 | `skill_heatmap` | Skill Heatmap & Competency Model / 技能热力图与胜任力模型 |
| 7 | `interest_assessment` | Career Interest Assessment / 职业兴趣评估 |
| 8 | `career_recommendations` | Career Recommendations / 基于兴趣的职业推荐 |
| 9 | `industry_analysis` | Industry Analysis / 行业分析 |
| 10 | `goal_setting` | Goal Setting / 目标设定 |
| 11 | `action_plan` | Action Plan / 行动计划 |
| 12 | `summary_outlook` | Summary & Outlook / 总结与展望 |

---

## Development 👩‍💻 开发

### Backend 后端

```bash
cd server

# Run in dev mode
go run ./cmd/server

# Build binary
go build -o bin/server ./cmd/server

# Run tests
go test ./...

# Lint
golangci-lint run
```

### Frontend 前端

```bash
cd web

# Install dependencies
npm install

# Dev mode with hot-reload (default: localhost:5173)
npm run dev

# Production build
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

- [ ] Report PDF export / 报告 PDF 导出功能
- [ ] Long-term conversation memory & summarization / 多轮对话记忆优化（长期记忆/摘要压缩）
- [ ] Front-end report visualization (radar chart / heatmap) / 前端报告可视化渲染（雷达图/热力图）
- [ ] In-app model switching / 多模型热切换（UI 中选择模型）
- [ ] Docker containerization / Docker 容器化部署
- [ ] i18n support / 多语言支持

---

## Contributing 🤝 贡献

Contributions are welcome! Please submit a Pull Request.

- Go code follows standard `gofmt` + `golangci-lint` conventions
- Front-end code follows the ESLint configuration
- Commit messages should use `feat:` / `fix:` / `docs:` / `refactor:` prefixes

---

## License 📄 许可证

[MIT License](LICENSE) © 2025 hautmz

---

## Acknowledgements 🙏 致谢

- [cloudwego/eino](https://github.com/cloudwego/eino) — ByteDance's open-source Go AI application framework
- [Gin](https://github.com/gin-gonic/gin) — Go HTTP web framework
- [Vue.js](https://vuejs.org/) — The progressive JavaScript framework
- [Element Plus](https://element-plus.org/) — Vue 3 component library
