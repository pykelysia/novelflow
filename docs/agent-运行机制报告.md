# NovelFlow Agent 运行机制报告

## 整体架构

NovelFlow 采用 **主 Agent + 3 个 Sub-Agent** 的分层协作结构，基于 CloudWeGo Eino ADK 的 `deep` 模式运行。

```
MainAgent
├── novel_outline_agent   (大纲规划)
├── novel_write_agent     (章节写作)
└── novel_review_agent    (质量审查)
```

主 Agent 不直接写入任何章节文件，只负责编排流程、审查质量、回复用户。所有正文写入由 `novel_write_agent` 完成。

---

## 1. Agent 初始化流程

`NewMainAgent` (`mainagent.go`) 是入口，依次完成：

1. 连接 MongoDB，创建或恢复 Session
2. 将 user-session 关联写入 MySQL（userID > 0 时）
3. 初始化三个 Sub-Agent（outline / write / review）
4. 加载 Skills 中间件（`getSkillsSystem`，从 `skills.base_dir` 读取）
5. 调用 `NewAgent` 创建底层 `adk.Runner`，挂载中间件

`NewAgent` (`agent.go`) 内部：

1. 调用 `getChatModel` 初始化主模型
2. 设置重试配置（最多 3 次，`isRetryAble` 判断）
3. 设置 `UnknownToolsHandler`
4. 构建 `SummarizationMiddleware`（Lite LLM）
5. 挂载 `SafeToolMiddleware` + `SummarizationMiddleware`
6. 系统提示中替换 `{session_id}` 和 `{user_rules}` 占位符

---

## 2. 会话管理（Session）

`session.go` 负责会话持久化：

- **存储**：MongoDB，`sessions` 集合存元数据，`messages` 集合存消息历史
- **新建/恢复**：传入长度不等于 36 的字符串创建新会话；传入合法 UUID（36字符）则从 MongoDB 恢复；若 UUID 不存在则自动新建
- **`Append()` 方法**：每条消息写入时携带 `session_id` 和 `created_at`，直接持久化到 MongoDB
- **`Use()` 方法**：按 `created_at` 升序从 MongoDB 读取本会话所有消息，重建为 Eino `adk.Message` 列表
  - `ContentType` → `schema.Message{Role, Content}`
  - `ToolResultType`（Assistant 角色）→ 重建 `ToolCalls`，ID 格式 `call_0`, `call_1`...
  - `ToolResultType`（Tool 角色）→ 重建 `ToolResult`，关联上一条消息的 ToolCallID
  - `ThinkingType` → **不参与历史重建**（仅流式输出，不写回 LLM 上下文）

---

## 3. 消息流（RunA）

`agent.go` 中的 `RunA` 方法处理流式输出：

```
用户消息 → Session.Append → Runner.Run → 逐步接收事件
    ├── MessageStream → 流式拼接内容 → handlerFunc(ContentType)
    ├── ReasoningContent → handlerFunc(ThinkingType)  [不保存到 Session]
    └── ToolName → handlerFunc(ToolType) → 保存工具调用记录 (ToolResultType)
```

输出最终拼接后保存回 Session，保证历史完整性。Thinking 内容只流式推送给调用方，不写入 MongoDB。

---

## 4. 中间件

### 主 Agent 中间件挂载顺序

`NewAgent` 内：`SafeToolMiddleware` → `SummarizationMiddleware`
`NewMainAgent` 追加：`SkillsMiddleware`

### Sub-Agent 中间件挂载顺序

`writer.go` / `outline.go` / `reviewer.go`：`SkillsMiddleware` → `SafeToolMiddleware`

| 中间件 | 挂载位置 | 作用 |
|---|---|---|
| `SafeToolMiddleware` | 主 Agent + 全部 Sub-Agent | 拦截工具调用错误，转换为提示字符串，`InterruptRerun` 错误除外 |
| `SummarizationMiddleware` | 主 Agent | 上下文超过 200k tokens 时，用 Lite LLM 自动摘要压缩历史 |
| `SkillsMiddleware` | 主 Agent + 全部 Sub-Agent | 从 `skills.base_dir` 目录加载写作技能模块，注入为可调用能力 |
| `UnknownToolsHandler` | 主 Agent + 全部 Sub-Agent | 对未注册工具调用返回 no-op 提示 |

### Sub-Agent 专属配置

所有三个 Sub-Agent 均设置：
- `WithoutGeneralSubAgent: true` — 禁止递归派生新 Sub-Agent
- `WithoutWriteTodos: true` — 禁止生成 TODO 列表

---

## 5. 双模型配置

| 模型 | 用途 | 配置键 |
|---|---|---|
| 主模型（Main LLM） | 全部 Agent（主 + 三个 Sub）的推理与生成 | `llm.*` |
| 轻量模型（Lite LLM） | 摘要中间件的上下文压缩；规则筛选 Chain | `lite_llm.*` |

支持三种后端，通过 `model_type` 配置切换：

| `model_type` | 后端实现 |
|---|---|
| `openai` | `eino-ext/components/model/openai` |
| `anthropic` | `eino-ext/components/model/claude` |
| `deepseek` | `eino-ext/components/model/deepseek` |

---

## 6. Sub-Agent 工具权限

工具名称均为实际注册的工具函数名：

| Agent | 可用工具 |
|---|---|
| `novel_outline_agent` | `write_outline_file_tool`, `read_outline_file_tool` |
| `novel_write_agent` | `write_novel_chapter_file_tool`, `edit_novel_chapter_file_tool`, `read_outline_file_tool`, `list_chapter_files_tool`, `read_novel_chapter_file_tool` |
| `novel_review_agent` | `list_chapter_files_tool`, `read_novel_chapter_file_tool`, `read_outline_file_tool`（只读，无写入） |
| 主 Agent | `read_novel_chapter_file_tool`（仅读取，写作任务委托给 Sub-Agent） |

章节文件存储路径：`{storage.novels_dir}/{sessionID}/{章节标题}.txt`
大纲文件路径：`{storage.novels_dir}/{sessionID}/outline.md`

---

## 7. 容错机制

- **LLM 429 重试**：最多 3 次，`isRetryAble` 判断是否可重试
- **工具错误隔离**：`SafeToolMiddleware` 将错误转为字符串（格式：`[tool error] ...`），会话不中断；`InterruptRerun` 错误透传不隔离
- **历史压缩**：`SummarizationMiddleware` 防止主 Agent 上下文超过 200k tokens
- **未知工具 no-op**：`defaultUnknownToolHandler` 对未注册工具返回提示，让模型重新选择
- **路径安全**：`readFileTool` / `editFileTool` 拒绝包含 `..` 的标题，防止目录穿越

---

## 8. 用户规则注入（Rules Injection）

每次生成前，`RunWithRules` (`workflow.go`) 在调用 `NewMainAgent` 之前执行规则筛选，将相关规则注入主 Agent 的系统提示。

### 整体流程

```
RunWithRules(ctx, sessionID, userID, enabledRules, intent)
        │
        ├─ enabledRules 为空 → rulesContent = ""
        │
        └─ buildRulesContent → runRuleSelectionChain
                │  任意 error → 降级注入全部规则
                │
                ▼
   compose.Chain[map[string]any, string]
   ┌─────────────────────────────────────────────┐
   │ AppendChatTemplate  map[string]any           │
   │   → []*schema.Message (System + User 消息)  │
   ├─────────────────────────────────────────────┤
   │ AppendChatModel (Lite LLM)                  │
   │   → *schema.Message (JSON 筛选结果)          │
   ├─────────────────────────────────────────────┤
   │ AppendLambda (parseSelectionResponse)        │
   │   → string (格式化后的 rulesContent)         │
   └─────────────────────────────────────────────┘
        │
        ▼
NewMainAgent(ctx, sessionID, userID, rulesContent)
```

### Chain 节点说明

| 节点 | 类型 | 输入 → 输出 | 职责 |
|---|---|---|---|
| ChatTemplate | `prompt.FromMessages(schema.FString, ...)` | `map[string]any` → `[]*schema.Message` | 将模板变量（genre/concept/style/requirements/rule_lines）渲染为 System + User 消息 |
| ChatModel | Lite LLM | `[]*schema.Message` → `*schema.Message` | 调用轻量模型，返回 `{"selected_ids": [...]}` JSON |
| Lambda | `parseSelectionResponse` | `*schema.Message` → `string` | 解析 JSON，查找 ID 对应规则，格式化为注入文本 |

### 降级策略

以下任一情况均触发降级，注入用户所有 enabled 规则：

- Lite LLM 初始化失败
- Chain 编译失败
- LLM 返回内容不含 `{...}` JSON 结构
- JSON 解析失败
- 筛选结果中所有 ID 均不在规则列表中（`selected` 为空）

降级逻辑收敛在 `buildRulesContent` 一处，`parseSelectionResponse` 内部的格式错误也统一降级而不返回 error。

### System Prompt 注入位置

主 Agent 系统提示末尾有 `{user_rules}` 占位符（`prompt.go`），`NewMainAgent` 在初始化时替换为筛选后的规则文本（无规则时替换为空字符串）。规则格式：

```
以下规则本次创作必须严格遵守（优先级高于一般写作建议）：

1. [规则名称] 规则内容
2. [规则名称] 规则内容
```
