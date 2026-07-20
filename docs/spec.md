# Spec: Eino Career Agent — AI 职业规划 Agent 应用

## Objective

用 Go + ByteDance Eino 框架重写 Career-Oracle，构建一个 AI 职业规划 Agent 应用。后端全栈 Go + SQLite，前端 Vue3 纯聊天界面。通过 Eino Agent 编排替代现有 4 种 menu_type 硬编码模式，实现单一聊天入口 + Agent 意图自动路由。

### 核心用户故事

1. 作为用户，我可以在一个聊天窗口中与 AI 对话，AI 自动识别我的意图（普通问答 / 生成报告 / 解析简历）并路由到对应处理流程
2. 作为用户，我可以说"帮我生成职业报告"，Agent 自动收集信息、并行生成 12 章节报告并返回
3. 作为用户，我可以上传简历文件（PDF/DOCX），Agent 自动解析内容并基于简历进行职业分析
4. 作为用户，我可以查看历史对话和已生成的报告
5. 作为管理员，我可以通过配置切换 LLM 提供商和模型

### Success Criteria

- 单一聊天入口，无需前端模式切换，Agent 自动意图路由准确率 > 90%
- 职业报告 12 章节并行生成，总耗时 < 单章节串行耗时的 1.5 倍
- 前端仅包含聊天 UI + 报告展示，无查询/管理类辅助页面
- 后端无 Python/Java/Redis/MongoDB 依赖，纯 Go + SQLite
- LLM 提供商可通过配置文件切换（OpenAI 兼容协议）
- 认证使用账密登录，无支付功能

## Tech Stack

### 后端
| 类别 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.22+ |
| AI 框架 | Eino (cloudwego/eino) | v0.9+ |
| Eino 扩展 | eino-ext (model/openai, tool, callbacks) | latest |
| Web 框架 | Gin | v1.10+ |
| 数据库 | SQLite (mattn/go-sqlite3 + gorm) | - |
| ORM | GORM | v1.25+ |
| 文件解析 | unidoc/unioffice (DOCX), pdfplumber-like Go lib (PDF) | - |
| JWT | golang-jwt/jwt | v5 |
| 配置 | viper + 环境变量 | OPENAI_API_KEY / OPENAI_BASE_URL / OPENAI_MODEL |
| 日志 | zap | - |
| SSE | gin-sse / custom SSE middleware | - |

### 前端
| 类别 | 技术 | 版本 |
|------|------|------|
| 框架 | Vue 3 | 3.4+ |
| 构建工具 | Vite | 5+ |
| UI 库 | Element Plus | - |
| 状态管理 | Pinia | - |
| HTTP | Axios | - |
| Markdown渲染 | markdown-it | - |
| SSE | eventsource / fetch SSE | - |

## Commands

```bash
# 后端
cd server && go build -o bin/server ./cmd/server       # 构建
cd server && go run ./cmd/server                         # 开发运行
cd server && go test ./...                               # 测试
cd server && golangci-lint run                           # Lint

# 前端
cd web && npm install                                    # 安装依赖
cd web && npm run dev                                    # 开发运行
cd web && npm run build                                  # 构建
cd web && npm run lint                                   # Lint

# 全项目
make dev                                                 # 同时启动前后端
make build                                               # 构建全部
```

## Project Structure

```
eino-carrer-agent/
├── docs/                           # 文档
│   └── spec.md                     # 本规格文档
├── server/                         # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # 程序入口
│   ├── internal/
│   │   ├── config/                 # 配置加载 (viper)
│   │   │   └── config.go
│   │   ├── domain/                 # 领域模型 / 实体定义
│   │   │   ├── user.go
│   │   │   ├── conversation.go
│   │   │   ├── message.go
│   │   │   └── report.go
│   │   ├── handler/                # HTTP Handler (Gin)
│   │   │   ├── chat.go             # 聊天 SSE 接口
│   │   │   ├── auth.go             # 登录/注册
│   │   │   ├── report.go           # 报告查询/导出
│   │   │   ├── upload.go           # 文件上传
│   │   │   └── middleware.go       # JWT 中间件
│   │   ├── repository/             # 数据访问层 (GORM)
│   │   │   ├── user_repo.go
│   │   │   ├── conversation_repo.go
│   │   │   ├── message_repo.go
│   │   │   └── report_repo.go
│   │   ├── service/                # 业务逻辑层
│   │   │   ├── auth_service.go
│   │   │   ├── chat_service.go
│   │   │   └── report_service.go
│   │   ├── agent/                  # Eino Agent 编排核心
│   │   │   ├── agent.go            # 主 Agent 构建（ChatModelAgent + Tools）
│   │   │   ├── tools/              # Agent Tools 定义
│   │   │   │   ├── report_tool.go      # 职业报告生成 Tool
│   │   │   │   ├── file_parse_tool.go  # 文件解析 Tool
│   │   │   │   └── intent_tool.go      # 意图识别辅助 Tool
│   │   │   ├── graph/              # Eino Graph 编排
│   │   │   │   ├── report_graph.go     # 报告 12 章节并行 Graph
│   │   │   │   └── chat_graph.go       # 普通对话 Graph
│   │   │   ├── prompts/            # Prompt 模板
│   │   │   │   ├── report_sections.go  # 12 章节各 prompt
│   │   │   │   ├── chat_prompt.go      # 对话 prompt
│   │   │   │   └── intent_prompt.go    # 意图路由 prompt
│   │   │   └── callback/          # Eino Callback（日志/追踪）
│   │   │       └── logging_callback.go
│   │   ├── parser/                 # 文件解析
│   │   │   ├── pdf_parser.go
│   │   │   └── docx_parser.go
│   │   └── pkg/                    # 内部公共工具
│   │       ├── jwt/
│   │       ├── sse/
│   │       └── response/
│   ├── configs/
│   │   └── config.yaml             # 通用配置（数据库路径、超时等，LLM 从环境变量读取）
│   ├── migrations/                 # SQLite 迁移脚本
│   ├── go.mod
│   └── go.sum
├── web/                            # Vue3 前端
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   ├── api/                    # API 调用封装
│   │   │   ├── chat.js
│   │   │   ├── auth.js
│   │   │   ├── report.js
│   │   │   └── upload.js
│   │   ├── components/             # 组件
│   │   │   ├── ChatWindow.vue          # 聊天窗口（核心）
│   │   │   ├── MessageBubble.vue       # 消息气泡
│   │   │   ├── ReportView.vue          # 报告展示
│   │   │   ├── FileUpload.vue          # 文件上传
│   │   │   └── LoginDialog.vue         # 登录弹窗
│   │   ├── stores/                 # Pinia 状态
│   │   │   ├── user.js
│   │   │   └── chat.js
│   │   ├── utils/
│   │   │   ├── request.js              # Axios 封装
│   │   │   ├── auth.js                 # Token 管理
│   │   │   └── sse.js                  # SSE 客户端
│   │   ├── views/
│   │   │   └── Chat.vue                # 唯一页面
│   │   └── assets/
│   ├── index.html
│   ├── vite.config.js
│   └── package.json
├── Makefile
└── README.md
```

## Code Style

### Go 后端风格

```go
// Package agent 提供 Eino Agent 编排核心逻辑
package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// ReportSectionInput 报告章节生成输入
type ReportSectionInput struct {
	ConversationHistory []*schema.Message // 用户对话历史
	SectionName         string            // 章节名称
	UserProfile         string            // 用户画像摘要
}

// ReportSectionOutput 报告章节生成输出
type ReportSectionOutput struct {
	SectionName string `json:"sectionName"` // 章节名称
	Content     string `json:"content"`     // 章节内容 JSON
	Err         error  `json:"-"`           // 生成错误
}

// NewReportGraph 构建报告 12 章节并行生成 Graph
// 使用 Eino Graph 编排确定性的并行结构，每个节点内部可使用 Agent+Tool
func NewReportGraph(ctx context.Context, chatModel model.BaseChatModel) (*compose.Graph[[]*schema.Message, *ReportResult], error) {
	graph := compose.NewGraph[[]*schema.Message, *ReportResult]()

	// 添加用户画像提取节点（串行前置）
	graph.AddLambdaNode("extract_profile", extractUserProfile)

	// 添加 12 个章节并行生成节点
	sections := GetReportSectionNames()
	for _, section := range sections {
		graph.AddLambdaNode(section, buildSectionGenerator(section, chatModel))
	}

	// 添加合并节点
	graph.AddLambdaNode("merge", mergeSections)

	// 串行：START → 提取画像
	graph.AddEdge(compose.START, "extract_profile")

	// 并行：画像提取 → 各章节
	for _, section := range sections {
		graph.AddEdge("extract_profile", section)
		graph.AddEdge(section, "merge")
	}

	// 串行：合并 → END
	graph.AddEdge("merge", compose.END)

	return graph, nil
}
```

### Vue3 前端风格

```vue
<!-- ChatWindow.vue - 聊天窗口核心组件 -->
<script setup>
import { ref, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { chatSSE } from '@/utils/sse'
import MessageBubble from './MessageBubble.vue'
import FileUpload from './FileUpload.vue'

const chatStore = useChatStore()
const inputMessage = ref('')
const isLoading = ref(false)

// 发送消息，通过 SSE 接收流式响应
async function sendMessage() {
  if (!inputMessage.value.trim() || isLoading.value) return

  const content = inputMessage.value.trim()
  inputMessage.value = ''
  isLoading.value = true

  chatStore.addMessage({ role: 'user', content })

  try {
    await chatSSE('/api/chat/stream', { content }, (event) => {
      // 处理 SSE 事件流
      chatStore.appendAssistantMessage(event.data)
    })
  } finally {
    isLoading.value = false
  }
}
</script>
```

### 关键约定

- Go: 结构体字段必须有 JSON tag，导出函数必须有注释
- Go: 错误处理不使用 panic，统一返回 error
- Go: 依赖注入通过构造函数参数，不使用全局变量
- Vue3: Composition API + `<script setup>` 语法
- Vue3: 所有 API 调用封装在 `api/` 目录
- 前端不直接操作 localStorage/sessionStorage，通过 store 封装
- 所有代码添加完整详细的中文注释

## Testing Strategy

### 后端测试

| 层级 | 框架 | 覆盖范围 | 位置 |
|------|------|----------|------|
| 单元测试 | Go testing | Agent Tool、Graph 节点、Service、Parser | `server/internal/**/*_test.go` |
| 集成测试 | Go testing + httptest | API Handler、SSE 流 | `server/internal/handler/*_test.go` |
| E2E测试 | Go testing | 完整 Agent 流程（mock LLM） | `server/test/e2e/` |

- 所有 Agent Tool 必须有单元测试
- Graph 编排必须有集成测试（验证并行结构正确性）
- LLM 调用在测试中 mock
- 报告生成超时 360 秒，通过 goroutine + SSE 推送进度
- 最低覆盖率目标：60%

### 前端测试

| 层级 | 框架 | 覆盖范围 |
|------|------|----------|
| 组件测试 | Vitest + Vue Test Utils | ChatWindow、ReportView |
| API Mock | msw | API 调用 |

## Boundaries

- **Always do**:
  - 添加完整详细的中文注释
  - 每个函数/方法必须有文档注释
  - LLM 调用配置化，不硬编码 API Key
  - SQLite migration 用版本管理
  - SSE 流式响应用于聊天
  - Agent Tool 声明清晰（Name、Description、Params）

- **Ask first**:
  - 引入新的第三方依赖
  - 修改数据库 Schema
  - 改变 Agent 编排结构（如从 Graph 改为纯 Agent）
  - 添加新的报告章节
  - 修改 Prompt 模板

- **Never do**:
  - 提交 API Key / Secret 到代码仓库
  - 使用 Python / Java / Node.js 在后端
  - 在前端实现业务逻辑（报告生成等）
  - 直接操作数据库而不通过 Repository 层
  - 使用全局可变状态

## API Design

### 认证

```
POST /api/auth/register    # 注册（账密）
POST /api/auth/login       # 登录，返回 JWT
POST /api/auth/logout      # 登出
```

### 聊天（核心）

```
POST /api/chat/stream      # SSE 流式聊天（统一入口）
  Request:  { content: string, file_ids?: string[] }
  Response: SSE event stream
    event: message          # 普通对话文本片段
    event: tool_call        # Agent 调用 Tool 通知
    event: report_progress  # 报告生成进度
    event: report_result    # 报告完整结果
    event: error            # 错误
    event: done             # 结束
```

### 文件上传

```
POST /api/upload           # 上传文件，返回 file_id
GET  /api/upload/:id       # 获取文件信息
```

### 报告

```
GET  /api/report/list      # 用户报告列表
GET  /api/report/:id       # 报告详情
```

### 对话管理

```
GET    /api/conversation/list       # 对话列表
GET    /api/conversation/:id        # 对话详情（含消息历史）
DELETE /api/conversation/:id        # 删除对话
```

### Agent 内部路由设计

前端不再区分 menu_type，所有请求走 `/api/chat/stream`。后端 Agent 自动意图路由：

```
用户消息 → ChatModelAgent（带 Tools）
           ├─ 意图：普通对话 → 直接 LLM 回复
           ├─ 意图：生成报告 → report_tool → Report Graph（12 章节并行）
           ├─ 意图：解析简历 → file_parse_tool → 解析后对话
           └─ 意图：查看报告 → report_query_tool → 返回已有报告
```

Agent 通过 Tool 的 Description 让 LLM 自主选择调用哪个 Tool，无需硬编码路由。

### Report Graph 内部结构

```
START → extract_profile（提取用户画像，串行前置）
            │
            ├─→ professional_index（专业指数评分）    ──┐
            ├─→ myself_report（个人信息提取）         ──┤
            ├─→ achievement_superiority（成就优势）    ──┤
            ├─→ career_experience（职业经历与成长路径）──┤
            ├─→ motivation_values（动机与价值观评估）  ──┤ 并行
            ├─→ skill_heatmap（技能热力图）           ──┤
            ├─→ interest_assessment（职业兴趣评估）    ──┤
            ├─→ career_recommendations（职业推荐）     ──┤
            ├─→ industry_analysis（行业分析）         ──┤
            ├─→ goal_setting（目标设定）              ──┤
            ├─→ action_plan（行动计划）               ──┤
            └─→ summary_outlook（总结与展望）         ──┘
                                                        │
            merge_sections（合并所有章节）←──────────────┘
                                                        │
END  ←────────────────────────────────────────────────┘
```

## Database Schema (SQLite)

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增ID |
| username | TEXT UNIQUE | 用户名 |
| password_hash | TEXT | 密码哈希（bcrypt） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### conversations
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | UUID |
| user_id | INTEGER FK | 用户ID |
| title | TEXT | 对话标题（自动生成） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### messages
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增ID |
| conversation_id | TEXT FK | 对话ID |
| role | TEXT | user/assistant/tool |
| content | TEXT | 消息内容 |
| tool_name | TEXT | Tool名称（role=tool时） |
| file_id | INTEGER FK | 关联文件ID |
| created_at | DATETIME | 创建时间 |

### reports
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | UUID |
| conversation_id | TEXT FK | 来源对话ID |
| user_id | INTEGER FK | 用户ID |
| professional_index | TEXT JSON | 专业指数评分 |
| myself_report | TEXT JSON | 个人信息 |
| achievement_superiority | TEXT JSON | 成就优势 |
| career_experience | TEXT JSON | 职业经历 |
| motivation_values | TEXT JSON | 动机与价值观 |
| skill_heatmap | TEXT JSON | 技能热力图 |
| interest_assessment | TEXT JSON | 职业兴趣 |
| career_recommendations | TEXT JSON | 职业推荐 |
| industry_analysis | TEXT JSON | 行业分析 |
| goal_setting | TEXT JSON | 目标设定 |
| action_plan | TEXT JSON | 行动计划 |
| summary_outlook | TEXT JSON | 总结展望 |
| status | TEXT | generating/completed/failed |
| created_at | DATETIME | 创建时间 |

### uploaded_files
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增ID |
| user_id | INTEGER FK | 用户ID |
| filename | TEXT | 原始文件名 |
| file_path | TEXT | 存储路径 |
| file_type | TEXT | pdf/docx |
| parsed_content | TEXT | 解析后文本内容 |
| file_size | INTEGER | 文件大小(bytes) |
| created_at | DATETIME | 创建时间 |

## Resolved Questions

1. **LLM 配置**：通过环境变量读取，`OPENAI_API_KEY`、`OPENAI_BASE_URL`、`OPENAI_MODEL`，无需在 config.yaml 中硬编码
2. **文件上传大小限制**：10MB
3. **报告生成超时**：360 秒
4. **后台任务队列**：先用 goroutine + SSE 推送进度，不引入消息队列
5. **前端渲染**：纯 SPA，不需要 SSR
6. **国际化**：仅中文，不需要 i18n
