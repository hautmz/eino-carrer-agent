# Tasks: Eino Career Agent 实现任务清单

基于 `plan.md` 的 6 阶段计划，拆分为可独立执行的原子任务。每个任务可在一个专注会话内完成。

---

## 阶段 1: 基础骨架

### Task 1.1: 初始化 Go 项目与目录结构
- **Acceptance**: `go build ./cmd/server` 编译通过，目录结构与 spec 一致
- **Verify**: `go build ./cmd/server` 无错误
- **Files**:
  - `server/go.mod`
  - `server/cmd/server/main.go` (空 main，仅 fmt.Println)
  - 创建所有 `internal/` 子目录（空 .gitkeep）

### Task 1.2: 添加核心依赖
- **Acceptance**: go.mod 包含 gin、gorm、sqlite、viper、zap、jwt 依赖，`go mod tidy` 成功
- **Verify**: `go mod tidy && go build ./cmd/server`
- **Files**: `server/go.mod`, `server/go.sum`

### Task 1.3: 配置系统
- **Acceptance**: config 结构体可从 config.yaml + 环境变量加载，环境变量优先级高于 yaml
- **Verify**: 单元测试覆盖环境变量覆盖逻辑
- **Files**:
  - `server/internal/config/config.go`
  - `server/configs/config.yaml`

### Task 1.4: 日志系统
- **Acceptance**: zap logger 初始化完成，支持开发(consle)/生产(json)两种格式
- **Verify**: 日志输出到 stdout 格式正确
- **Files**: `server/internal/pkg/logger/logger.go`

### Task 1.5: 程序入口 + 健康检查
- **Acceptance**: `go run ./cmd/server` 启动后 `curl /health` 返回 200
- **Verify**: `curl localhost:8080/health` → `{"success":true,"message":"OK"}`
- **Files**: `server/cmd/server/main.go`

---

## 阶段 2: 数据层

### Task 2.1: GORM 数据库初始化 + AutoMigrate
- **Acceptance**: 启动时自动创建 SQLite 文件及所有表，表结构符合 spec 定义
- **Verify**: 启动后 SQLite 文件存在，`sqlite3 data/eino_career.db .tables` 列出 5 张表
- **Files**:
  - `server/internal/domain/user.go`
  - `server/internal/domain/conversation.go`
  - `server/internal/domain/message.go`
  - `server/internal/domain/report.go`
  - `server/internal/domain/uploaded_file.go`
  - `server/internal/pkg/database/database.go`

### Task 2.2: 统一响应 + JWT 工具
- **Acceptance**: Response 可序列化为 `{"success":true,"message":"OK","data":null,"errors":{}}`；JWT 可生成、解析、验证
- **Verify**: 单元测试通过
- **Files**:
  - `server/internal/pkg/response/response.go`
  - `server/internal/pkg/jwt/jwt.go`

### Task 2.3: User Repository
- **Acceptance**: CRUD + 按用户名查询，接口定义与实现分离
- **Verify**: `go test ./internal/repository/user_repo_test.go`
- **Files**: `server/internal/repository/user_repo.go`

### Task 2.4: Conversation Repository
- **Acceptance**: CRUD + 按用户ID分页列表
- **Verify**: `go test ./internal/repository/conversation_repo_test.go`
- **Files**: `server/internal/repository/conversation_repo.go`

### Task 2.5: Message Repository
- **Acceptance**: 批量插入 + 按对话ID分页查询（按时间正序）
- **Verify**: `go test ./internal/repository/message_repo_test.go`
- **Files**: `server/internal/repository/message_repo.go`

### Task 2.6: Report Repository
- **Acceptance**: CRUD + 按用户ID列表 + 状态更新方法
- **Verify**: `go test ./internal/repository/report_repo_test.go`
- **Files**: `server/internal/repository/report_repo.go`

### Task 2.7: UploadedFile Repository
- **Acceptance**: CRUD + 解析内容更新方法
- **Verify**: `go test ./internal/repository/uploaded_file_repo_test.go`
- **Files**: `server/internal/repository/uploaded_file_repo.go`

---

## 阶段 3: Agent 核心

### Task 3.1: ChatModel 初始化
- **Acceptance**: 根据环境变量 OPENAI_API_KEY/OPENAI_BASE_URL/OPENAI_MODEL 创建 eino-ext OpenAI ChatModel，可调用 Generate
- **Verify**: 单元测试（mock HTTP）验证配置传递正确
- **Files**: `server/internal/agent/model.go`

### Task 3.2: 对话 System Prompt
- **Acceptance**: 职业规划师角色 Prompt 定义完整，包含意图路由说明（何时调用 report_tool 等）
- **Verify**: Prompt 常量可被其他包引用
- **Files**: `server/internal/agent/prompts/chat_prompt.go`

### Task 3.3: 报告章节 Prompt 模板（12 个）
- **Acceptance**: 12 个章节各自的 Prompt 模板定义，每个 Prompt 清晰描述输入输出格式（JSON Schema）
- **Verify**: 编译通过，Prompt 常量可被引用
- **Files**: `server/internal/agent/prompts/report_sections.go`

### Task 3.4: 用户画像提取 Prompt
- **Acceptance**: 从对话历史提取用户画像的 Prompt 定义，输出结构化文本
- **Verify**: 编译通过
- **Files**: `server/internal/agent/prompts/profile_prompt.go`

### Task 3.5: Report Graph — 基础结构
- **Acceptance**: Graph 定义完成：START → extract_profile → (12 章节并行) → merge → END，可 Compile
- **Verify**: `go test ./internal/agent/graph/report_graph_test.go` 验证节点和边关系
- **Files**: `server/internal/agent/graph/report_graph.go`

### Task 3.6: Report Graph — 章节生成节点实现
- **Acceptance**: 每个章节节点内部调用 ChatModel + 对应 Prompt，解析 JSON 响应，带超时和错误处理
- **Verify**: 单个章节节点单元测试（mock ChatModel）
- **Files**: `server/internal/agent/graph/report_graph.go` (补充节点实现)

### Task 3.7: Report Graph — 并发控制与合并
- **Acceptance**: semaphore 限制最多 4 路并行；merge 节点正确收集 12 个章节结果；总超时 360s
- **Verify**: 集成测试验证并行执行和合并逻辑
- **Files**: `server/internal/agent/graph/report_graph.go` (补充并发控制和 merge)

### Task 3.8: report_tool — 职业报告生成 Tool
- **Acceptance**: ToolInfo（name=generate_career_report, desc, params）定义正确；Invoke 触发 Report Graph 并返回结果
- **Verify**: 单元测试验证 ToolInfo 和 Invoke 逻辑
- **Files**: `server/internal/agent/tools/report_tool.go`

### Task 3.9: file_parse_tool — 文件解析 Tool
- **Acceptance**: ToolInfo 定义正确；Invoke 接收 file_id，从 DB 读取文件、调用 parser、返回文本内容
- **Verify**: 单元测试（mock repo + parser）
- **Files**: `server/internal/agent/tools/file_parse_tool.go`

### Task 3.10: report_query_tool — 报告查询 Tool
- **Acceptance**: ToolInfo 定义正确；Invoke 接收可选 report_id，查询 DB 返回报告摘要
- **Verify**: 单元测试（mock repo）
- **Files**: `server/internal/agent/tools/report_query_tool.go`

### Task 3.11: PDF 解析器
- **Acceptance**: 输入 PDF 文件路径，输出提取的文本内容；能处理中文 PDF
- **Verify**: 用测试 PDF 文件验证提取结果
- **Files**: `server/internal/parser/pdf_parser.go`

### Task 3.12: DOCX 解析器
- **Acceptance**: 输入 DOCX 文件路径，输出提取的文本内容
- **Verify**: 用测试 DOCX 文件验证提取结果
- **Files**: `server/internal/parser/docx_parser.go`

### Task 3.13: 主 Agent 构建
- **Acceptance**: 使用 Eino ADK ChatModelAgent 构建，注册 System Prompt + 3 个 Tools + Callback；可 Invoke 对话
- **Verify**: 集成测试（mock ChatModel）验证 Agent 可响应并选择 Tool
- **Files**: `server/internal/agent/agent.go`

### Task 3.14: Agent Callback 日志
- **Acceptance**: 实现 Eino Callback Handler 接口，OnStart/OnEnd/OnError 记录日志
- **Verify**: Agent 运行时日志输出正确
- **Files**: `server/internal/agent/callback/logging_callback.go`

---

## 阶段 4: API 层

### Task 4.1: Auth Handler（注册/登录）
- **Acceptance**: POST /api/auth/register 创建用户（密码 bcrypt）；POST /api/auth/login 验证密码返回 JWT
- **Verify**: httptest 集成测试
- **Files**:
  - `server/internal/handler/auth.go`
  - `server/internal/service/auth_service.go`

### Task 4.2: JWT 认证中间件
- **Acceptance**: 从 Authorization: Bearer {token} 解析 userID 注入 context；无效 token 返回 401
- **Verify**: httptest 测试有效/无效/缺失 token 场景
- **Files**: `server/internal/handler/middleware.go`

### Task 4.3: SSE 工具函数
- **Acceptance**: 提供 WriteSSE(writer, event, data) 函数，支持事件类型 + 数据写入 + 心跳
- **Verify**: 单元测试验证 SSE 格式正确（`event: xxx\ndata: xxx\n\n`）
- **Files**: `server/internal/pkg/sse/sse.go`

### Task 4.4: Chat Handler — SSE 流式聊天
- **Acceptance**: POST /api/chat/stream 返回 SSE 事件流；正确处理 message/tool_call/error/done 事件；心跳保活 15s
- **Verify**: httptest + SSE 客户端测试验证事件流格式
- **Files**:
  - `server/internal/handler/chat.go`
  - `server/internal/service/chat_service.go`

### Task 4.5: Chat Service — 对话上下文管理
- **Acceptance**: 从 DB 加载历史消息转为 Eino Message 格式；用户消息先存 DB；assistant 回复拼接后存 DB；新建/续用对话
- **Verify**: 单元测试验证消息格式转换和持久化
- **Files**: `server/internal/service/chat_service.go` (补充)

### Task 4.6: Chat Service — 报告生成 SSE 推送
- **Acceptance**: report_tool 触发时，goroutine 启动 Report Graph，通过 channel 推送 report_progress 事件，完成后推送 report_result 事件
- **Verify**: 集成测试验证 SSE 事件序列（progress → result → done）
- **Files**: `server/internal/service/chat_service.go` (补充), `server/internal/service/report_service.go`

### Task 4.7: Upload Handler
- **Acceptance**: POST /api/upload 接收文件，限 10MB，仅 PDF/DOCX；上传后异步解析存 parsed_content；GET /api/upload/:id 返回文件信息
- **Verify**: httptest 上传测试文件 + 查询文件信息
- **Files**:
  - `server/internal/handler/upload.go`
  - `server/internal/service/upload_service.go`

### Task 4.8: Report Handler
- **Acceptance**: GET /api/report/list 分页列表；GET /api/report/:id 返回完整 12 章节 JSON
- **Verify**: httptest 测试
- **Files**: `server/internal/handler/report.go`

### Task 4.9: Conversation Handler
- **Acceptance**: GET /api/conversation/list + GET /api/conversation/:id + DELETE /api/conversation/:id
- **Verify**: httptest 测试
- **Files**: `server/internal/handler/conversation.go`

### Task 4.10: 路由注册 + CORS
- **Acceptance**: 所有路由注册完成；公开/认证路由分组正确；CORS 允许前端开发域
- **Verify**: `go run ./cmd/server` 启动后 `curl` 各路由返回正确状态码
- **Files**: `server/cmd/server/main.go` 或 `server/internal/router/router.go`

---

## 阶段 5: 前端

### Task 5.1: Vue3 项目初始化
- **Acceptance**: `npm run dev` 启动成功；Vite 代理 /api → localhost:8080；Element Plus 可用
- **Verify**: 浏览器访问显示空页面无报错
- **Files**:
  - `web/package.json`
  - `web/vite.config.js`
  - `web/src/main.js`
  - `web/src/App.vue`
  - `web/index.html`

### Task 5.2: Axios 封装 + Token 管理
- **Acceptance**: request 实例自动附加 JWT；401 时跳转登录；auth.js 管理 token 存储
- **Verify**: 手动测试或组件测试
- **Files**:
  - `web/src/utils/request.js`
  - `web/src/utils/auth.js`

### Task 5.3: SSE 客户端封装
- **Acceptance**: chatSSE 函数支持按 event type 分发回调；支持心跳忽略；断线重连
- **Verify**: 连接后端 SSE 接口测试
- **Files**: `web/src/utils/sse.js`

### Task 5.4: API 封装
- **Acceptance**: auth/chat/upload/report/conversation API 函数全部定义
- **Verify**: 导入无报错
- **Files**:
  - `web/src/api/auth.js`
  - `web/src/api/chat.js`
  - `web/src/api/upload.js`
  - `web/src/api/report.js`
  - `web/src/api/conversation.js`

### Task 5.5: Pinia Stores
- **Acceptance**: user store（认证状态）+ chat store（对话/消息/SSE 状态）功能完整
- **Verify**: 组件内使用 store 无报错
- **Files**:
  - `web/src/stores/user.js`
  - `web/src/stores/chat.js`

### Task 5.6: LoginDialog 组件
- **Acceptance**: 登录/注册表单切换；表单验证；调用 API 后更新 store；登录成功自动关闭
- **Verify**: 手动测试
- **Files**: `web/src/components/LoginDialog.vue`

### Task 5.7: MessageBubble 组件
- **Acceptance**: 区分 user/assistant/tool 三种角色样式；assistant 消息支持 markdown 渲染
- **Verify**: 渲染测试消息样式正确
- **Files**: `web/src/components/MessageBubble.vue`

### Task 5.8: FileUpload 组件
- **Acceptance**: el-upload 封装；限制 10MB + PDF/DOCX；上传成功返回 file_id 附加到消息
- **Verify**: 上传测试文件成功
- **Files**: `web/src/components/FileUpload.vue`

### Task 5.9: ReportView 组件
- **Acceptance**: 12 章节卡片展示；每章节可折叠/展开；数据可视化（进度条/表格）；加载状态
- **Verify**: 传入 mock 报告数据渲染正确
- **Files**: `web/src/components/ReportView.vue`

### Task 5.10: ChatWindow 组件
- **Acceptance**: 消息列表滚动到底部；输入框 + 发送；SSE 流式拼接 assistant 消息；Tool Call 提示；报告进度条
- **Verify**: 连接后端完整对话流程
- **Files**: `web/src/components/ChatWindow.vue`

### Task 5.11: Chat 主页面布局
- **Acceptance**: 左侧对话列表 + 右侧聊天窗口；新建对话；切换对话；删除对话
- **Verify**: 完整页面交互正常
- **Files**: `web/src/views/Chat.vue`

---

## 阶段 6: 集成联调

### Task 6.1: Makefile
- **Acceptance**: `make dev` 并行启动前后端；`make build` 构建前后端产物
- **Verify**: `make dev` 启动后可访问完整应用
- **Files**: `Makefile`

### Task 6.2: 前端构建产物集成到 Go 服务
- **Acceptance**: Gin 服务 static 目录托管前端 SPA；所有非 /api 路由返回 index.html
- **Verify**: `make build && ./bin/server` 单二进制启动后前端可用
- **Files**: `server/cmd/server/main.go` (补充 static 服务)

### Task 6.3: 端到端功能验证
- **Acceptance**: 注册→登录→普通对话→上传简历→生成报告→查看报告→切换对话 全流程通过
- **Verify**: 手动全流程操作
- **Files**: 可能修复少量 bug，无新文件

### Task 6.4: 边界情况修复
- **Acceptance**: SSE 断线重连、报告超时提示、文件格式/大小错误提示、空消息拦截均正常
- **Verify**: 手动触发各边界场景
- **Files**: 按需修复

---

## 执行顺序与依赖

```
1.1 → 1.2 → 1.3 → 1.4 → 1.5 → 2.1 → 2.2 → 2.3~2.7(可并行) →
3.1 → 3.2~3.4(可并行) → 3.5 → 3.6 → 3.7 → 3.8~3.10(可并行) →
3.11~3.12(可并行) → 3.13 → 3.14 →
4.1~4.2(可并行) → 4.3 → 4.4 → 4.5 → 4.6 → 4.7~4.9(可并行) → 4.10 →
5.1 → 5.2~5.4(可并行) → 5.5 → 5.6~5.9(可并行) → 5.10 → 5.11 →
6.1 → 6.2 → 6.3 → 6.4
```

**关键路径**: 1.x → 2.x → 3.x → 4.x → 6.x
**并行机会**: 阶段 4 和阶段 5 可并行开发（API 契约已定义）
