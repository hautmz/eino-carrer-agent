# Plan: Eino Career Agent 技术实现计划

## 总览

基于 spec.md 的规格定义，按依赖关系将实现分为 6 个阶段，每阶段有明确的验证检查点。

### 实现顺序总图

```
阶段1: 基础骨架 ──→ 阶段2: 数据层 ──→ 阶段3: Agent核心 ──→ 阶段4: API层 ──→ 阶段5: 前端 ──→ 阶段6: 集成联调
  (项目初始化)       (Schema+Repo)     (Eino编排+Tools)     (Handler+SSE)      (Vue3聊天)      (端到端验证)
```

### 风险与缓解

| 风险 | 影响 | 缓解策略 |
|------|------|----------|
| Eino ADK API 不稳定（v0.9 快速迭代中） | Agent 编排代码可能需要调整 | 先用 Graph + compose 包（更稳定），ADK 作为上层封装 |
| 12 章节并行 LLM 调用 token 消耗大 | API 成本高、可能触发限流 | 限制并发数（最多 4 路），加上重试和超时 |
| SQLite 并发写入锁 | 报告并行写入可能冲突 | 12 章节结果先在内存合并，一次性写入 |
| SSE 长连接断开 | 报告生成 360s 超时内可能断连 | 心跳保活 + 前端断线重连 + 报告状态可查询 |

---

## 阶段 1: 基础骨架（项目初始化）

**目标**: Go 项目初始化、目录结构、配置加载、日志，能编译运行空服务

### 1.1 初始化 Go Module + 目录结构

- 创建 `server/` 下完整目录结构（cmd/internal/configs/migrations）
- `go mod init github.com/user/eino-carrer-agent/server`
- 添加核心依赖：gin、gorm、sqlite driver、viper、zap、golang-jwt

### 1.2 配置系统 (config)

- `internal/config/config.go`：用 viper 加载 config.yaml + 环境变量
- `configs/config.yaml`：数据库路径、服务端口、超时时间、文件上传限制
- 环境变量覆盖：`OPENAI_API_KEY`、`OPENAI_BASE_URL`、`OPENAI_MODEL`
- 配置结构体定义

### 1.3 日志系统

- `internal/pkg/logger/`：基于 zap 的日志初始化
- 支持开发/生产两种日志格式

### 1.4 程序入口 + 空路由

- `cmd/server/main.go`：加载配置 → 初始化日志 → 初始化数据库 → 启动 Gin
- 注册健康检查路由 `GET /health`

### 验证检查点
- `go build ./cmd/server` 编译通过
- `go run ./cmd/server` 启动后 `curl localhost:8080/health` 返回 200

---

## 阶段 2: 数据层（Schema + Repository）

**目标**: SQLite 数据库建表、GORM 模型、Repository 层 CRUD

### 2.1 数据库初始化 + AutoMigrate

- `internal/domain/`：定义所有 GORM 模型（User, Conversation, Message, Report, UploadedFile）
- 数据库初始化逻辑，AutoMigrate 自动建表
- SQLite 文件路径从配置读取，默认 `./data/eino_career.db`

### 2.2 Repository 层

- `internal/repository/user_repo.go`：用户 CRUD + 按用户名查询
- `internal/repository/conversation_repo.go`：对话 CRUD + 按用户ID分页列表
- `internal/repository/message_repo.go`：消息批量插入 + 按对话ID分页查询
- `internal/repository/report_repo.go`：报告 CRUD + 按用户ID列表 + 状态更新
- 每个 Repo 接口定义 + 实现，方便后续 mock 测试

### 2.3 公共工具

- `internal/pkg/response/`：统一 API 响应格式 `{success, message, data, errors}`
- `internal/pkg/jwt/`：JWT 生成 + 解析 + 中间件

### 验证检查点
- 单元测试：每个 Repository 的 CRUD 操作通过
- `go test ./internal/repository/...` 全部 PASS

---

## 阶段 3: Agent 核心（Eino 编排 + Tools）

**目标**: Eino ChatModel 初始化、主 Agent 构建、Report Graph 并行编排、Agent Tools 定义

### 3.1 ChatModel 初始化

- `internal/agent/model.go`：根据环境变量创建 OpenAI 兼容 ChatModel
- 使用 `eino-ext/components/model/openai` 包
- 支持自定义 BaseURL（兼容 Qwen/DeepSeek 等）
- Model failover 可选配置

### 3.2 Prompt 模板

- `internal/agent/prompts/chat_prompt.go`：职业规划师 System Prompt
- `internal/agent/prompts/intent_prompt.go`：意图路由辅助 Prompt
- `internal/agent/prompts/report_sections.go`：12 个章节各自的 Prompt 模板
  - professional_index：专业指数评分
  - myself_report：个人信息提取
  - achievement_superiority：成就与优势
  - career_experience：职业经历与成长路径
  - motivation_values：动机与价值观评估
  - skill_heatmap：技能热力图与胜任力模型
  - interest_assessment：职业兴趣评估
  - career_recommendations：基于兴趣的职业推荐
  - industry_analysis：行业分析
  - goal_setting：目标设定
  - action_plan：行动计划
  - summary_outlook：总结与展望
- `internal/agent/prompts/profile_prompt.go`：用户画像提取 Prompt

### 3.3 Report Graph（核心）

- `internal/agent/graph/report_graph.go`：
  - 构建并行 Graph：START → extract_profile → (12 章节并行) → merge → END
  - extract_profile 节点：从对话历史中提取用户画像摘要
  - 12 个章节节点：每个章节是一个 Lambda，内部调用 ChatModel + 对应 Prompt
  - merge 节点：合并所有章节结果为 ReportResult
  - Graph Compile + Invoke/Stream 调用
- 并发限制：使用 semaphore 控制最多 4 路并行 LLM 调用
- 超时控制：单章节 120s，总报告 360s（通过 context.WithTimeout）

### 3.4 Agent Tools 定义

- `internal/agent/tools/report_tool.go`：
  - Tool 名称：`generate_career_report`
  - 描述：当用户请求生成职业规划报告时调用，需要先通过对话收集足够信息
  - 参数：无（从对话历史自动获取）
  - 实现：触发 Report Graph，通过 channel 推送进度
- `internal/agent/tools/file_parse_tool.go`：
  - Tool 名称：`parse_resume_file`
  - 描述：解析用户上传的简历文件（PDF/DOCX），提取文本内容用于职业分析
  - 参数：file_id (string)
  - 实现：从数据库读取文件 → 解析 → 返回文本内容注入对话
- `internal/agent/tools/report_query_tool.go`：
  - Tool 名称：`query_career_report`
  - 描述：查询用户已生成的职业规划报告
  - 参数：report_id (string, optional)
  - 实现：从 SQLite 查询报告并返回摘要

### 3.5 主 Agent 构建

- `internal/agent/agent.go`：
  - 使用 Eino ADK `ChatModelAgent` 构建主 Agent
  - 配置 System Prompt（职业规划师角色 + 意图路由说明）
  - 注册 3 个 Tools：generate_career_report、parse_resume_file、query_career_report
  - 配置 Callback（日志追踪）
- `internal/agent/callback/logging_callback.go`：基于 Eino Callback 接口的日志记录

### 3.6 文件解析器

- `internal/parser/pdf_parser.go`：PDF 文本提取（使用 ledongthuh/pdfplumber 或类似 Go 库）
- `internal/parser/docx_parser.go`：DOCX 文本提取（使用 nguyenthenguyen/docx 或 unioffice）

### 验证检查点
- 单元测试：每个 Tool 的 ToolInfo 定义正确
- 单元测试：Report Graph 的节点连接关系正确（可 mock ChatModel）
- 单元测试：文件解析器能正确提取 PDF/DOCX 文本（用测试文件）
- `go test ./internal/agent/...` 全部 PASS

---

## 阶段 4: API 层（Handler + SSE）

**目标**: 完整 HTTP API、SSE 流式推送、JWT 认证中间件、文件上传

### 4.1 认证 Handler

- `internal/handler/auth.go`：
  - `POST /api/auth/register`：注册（用户名 + 密码，bcrypt 哈希）
  - `POST /api/auth/login`：登录验证，返回 JWT token
- `internal/handler/middleware.go`：JWT 认证中间件，解析 token 注入 userID

### 4.2 聊天 Handler（核心）

- `internal/handler/chat.go`：
  - `POST /api/chat/stream`：SSE 流式聊天
  - 请求体：`{ content, conversation_id?, file_ids? }`
  - 无 conversation_id 则新建对话
  - SSE 事件类型：
    - `message`：普通对话文本片段（从 Agent Stream 迭代器读取）
    - `tool_call`：Agent 调用了哪个 Tool
    - `report_progress`：报告生成进度（如 "3/12 章节已完成"）
    - `report_result`：报告完整 JSON 结果
    - `error`：错误信息
    - `done`：结束标记
  - 心跳保活：每 15s 发送 `:heartbeat` 注释
- `internal/service/chat_service.go`：
  - 管理对话上下文（从 DB 加载历史消息 → 转为 Eino Message 格式）
  - 调用 Agent Run，将迭代器输出转为 SSE 事件
  - 报告生成时：goroutine 启动 Report Graph，通过 channel 推送进度
  - 消息持久化：用户消息先存 DB，Agent 回复流式拼接完成后存 DB
- `internal/pkg/sse/`：SSE 写入工具函数

### 4.3 文件上传 Handler

- `internal/handler/upload.go`：
  - `POST /api/upload`：文件上传，限制 10MB，支持 PDF/DOCX
  - 上传后异步解析文本内容存入 parsed_content 字段
  - `GET /api/upload/:id`：查询文件信息

### 4.4 报告 Handler

- `internal/handler/report.go`：
  - `GET /api/report/list`：用户报告列表（分页）
  - `GET /api/report/:id`：报告详情（含 12 章节完整 JSON）

### 4.5 对话 Handler

- `internal/handler/conversation.go`：
  - `GET /api/conversation/list`：对话列表
  - `GET /api/conversation/:id`：对话详情含消息历史
  - `DELETE /api/conversation/:id`：删除对话

### 4.6 路由注册

- `cmd/server/main.go` 或独立 router 文件：
  - 公开路由：/health、/api/auth/*
  - 认证路由：/api/chat/*、/api/upload/*、/api/report/*、/api/conversation/*
  - CORS 中间件

### 验证检查点
- 集成测试：注册 → 登录 → 获取 token → 调用 /api/chat/stream 收到 SSE 事件
- 集成测试：上传文件 → 文件解析 → file_id 有效
- 集成测试：对话历史持久化和查询
- `go test ./internal/handler/...` 全部 PASS

---

## 阶段 5: 前端（Vue3 聊天界面）

**目标**: 纯聊天 SPA，单一页面，支持 SSE 流式对话、文件上传、报告展示

### 5.1 项目初始化

- `npm create vue@latest` 创建 Vue3 + Vite 项目
- 安装依赖：element-plus、pinia、axios、markdown-it、vue-router
- 配置 vite.config.js：代理 /api → localhost:8080

### 5.2 API 封装

- `src/api/auth.js`：register、login
- `src/api/chat.js`：chatSSE（fetch SSE 封装）
- `src/api/upload.js`：上传文件
- `src/api/report.js`：报告列表、报告详情
- `src/api/conversation.js`：对话列表、对话详情、删除
- `src/utils/request.js`：Axios 实例 + JWT 拦截器
- `src/utils/sse.js`：SSE 客户端封装（支持事件类型分发、断线重连）

### 5.3 状态管理

- `src/stores/user.js`：用户认证状态、token、登录/登出
- `src/stores/chat.js`：
  - 当前对话列表
  - 当前对话消息
  - 正在生成的报告
  - SSE 连接状态
  - 方法：sendMessage、loadHistory、createConversation

### 5.4 核心组件

- `src/views/Chat.vue`：唯一页面，布局为左侧对话列表 + 右侧聊天窗口
- `src/components/ChatWindow.vue`：
  - 消息列表渲染（用户/assistant 气泡）
  - 输入框 + 发送按钮
  - SSE 流式接收时实时拼接 assistant 消息
  - Tool Call 提示（如 "正在生成职业报告..."）
  - 报告进度条显示
- `src/components/MessageBubble.vue`：
  - 用户消息：右侧气泡
  - Assistant 消息：左侧气泡 + markdown 渲染
  - Tool 消息：特殊样式提示
- `src/components/ReportView.vue`：
  - 报告 12 章节卡片式展示
  - 每个章节可折叠/展开
  - 技能热力图等数据可视化（简单表格/进度条即可，不做复杂图表）
- `src/components/FileUpload.vue`：
  - Element Plus el-upload 封装
  - 限制 10MB、PDF/DOCX
  - 上传成功返回 file_id 附加到消息
- `src/components/LoginDialog.vue`：
  - 登录/注册表单弹窗
  - 未登录时自动弹出

### 5.5 样式

- 整体暗色主题，专业简洁
- 聊天气泡区分用户/AI
- 报告展示用卡片布局
- 响应式设计（移动端适配）

### 验证检查点
- `npm run build` 构建成功
- 手动验证：登录 → 发消息 → 收到流式回复 → 上传文件 → 生成报告 → 查看报告

---

## 阶段 6: 集成联调

**目标**: 前后端联调、端到端功能验证

### 6.1 联调清单

- 注册/登录流程
- 普通对话（SSE 流式）
- 文件上传 + 简历解析对话
- 触发报告生成（Agent 意图识别 → report_tool → Graph 并行 → SSE 推送进度 → 完整结果）
- 报告查看（点击已生成报告展示 12 章节）
- 对话历史查看和切换
- 对话删除

### 6.2 边界情况

- 网络断开后 SSE 重连
- 报告生成超时（360s）后的错误提示
- 文件上传格式/大小错误
- 并发对话场景
- 空/超长消息处理

### 6.3 部署准备

- Makefile：dev（并行启动前后端）、build（构建前后端）
- 前端构建产物放入 server 的 static 目录供 Gin 直接服务
- 单二进制 + SQLite 文件 + static 目录 = 完整部署

### 验证检查点
- 全流程端到端测试通过
- `make build` 生成可部署产物
- 单二进制启动后前端可正常访问

---

## 依赖关系与并行机会

```
阶段1 ──→ 阶段2 ──→ 阶段3 ──→ 阶段4 ──→ 阶段6
                                    ↗
                              阶段5 ──┘
```

- 阶段 5（前端）在阶段 3 完成后可并行开发（API 契约已定义）
- 阶段 4 和阶段 5 可并行
- 阶段 3 是关键路径，最复杂的部分（Eino Agent 编排）
