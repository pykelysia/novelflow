# NovelFlow Agent 运行机制报告

## 整体架构

NovelFlow 采用 **主 Agent + 3 个 Sub-Agent** 的分层协作结构，基于 CloudWeGo Eino ADK 的 `deep` 模式运行。

```
MainAgent
├── novel_outline_agent   (大纲规划)
├── novel_write_agent     (章节写作)
└── novel_review_agent    (质量审查)
```

---

## 1. Agent 初始化流程

`NewMainAgent` (`mainagent.go`) 是入口，依次完成：

1. 连接 MongoDB，创建或恢复 Session
2. 将 user-session 关联写入 MySQL（userID > 0 时）
3. 初始化三个 Sub-Agent（outline / write / review）
4. 加载 `.skills/` 目录下的写作技能模块（Eino Skill Middleware）
5. 调用 `NewAgent` 创建底层 `adk.Runner`，挂载中间件

---

## 2. 会话管理（Session）

`session.go` 负责会话持久化：

- **存储**：MongoDB，`sessions` 集合存元数据，`messages` 集合存消息历史
- **新建/恢复**：传入空字符串或非 UUID 字符串创建新会话；传入合法 UUID 则从 MongoDB 恢复
- **`Use()` 方法**：将存储的消息重建为 Eino `adk.Message` 列表，包含工具调用的 ID 重组逻辑（`call_0`, `call_1`...）
- **消息类型**：`Content`、`Thinking`、`Tool`、`ToolResult` 四种，按类型差异化处理

---

## 3. 消息流（RunA）

`agent.go` 中的 `RunA` 方法处理流式输出：

```
用户消息 → Session.Append → Runner.Run → 逐步接收事件
    ├── MessageStream → 流式拼接内容 → handlerFunc(ContentType)
    ├── ReasoningContent → handlerFunc(ThinkingType)
    └── ToolName → handlerFunc(ToolType) → 保存工具调用记录
```

输出最终拼接后保存回 Session，保证历史完整性。

---

## 4. 中间件

挂载顺序：`SafeToolMiddleware` → `SummarizationMiddleware`（以及 Skills Middleware）

| 中间件 | 作用 |
|---|---|
| `SafeToolMiddleware` | 拦截工具调用错误，转换为提示字符串而非中断流程，`InterruptRerun` 错误除外 |
| `SummarizationMiddleware` | 上下文超过 200k tokens 时，用 Lite LLM 自动摘要压缩历史 |
| `SkillsMiddleware` | 从 `.skills/` 目录加载写作技能模块，注入为可调用能力 |
| `UnknownToolsHandler` | 对未注册工具调用返回 no-op 提示，防止 Agent 崩溃 |

---

## 5. 双模型配置

| 模型 | 用途 | 配置键 |
|---|---|---|
| 主模型（Main LLM） | 所有 Agent 的推理与生成 | `llm.*` |
| 轻量模型（Lite LLM） | 摘要中间件的上下文压缩；用户规则筛选 Chain | `lite_llm.*` |

支持三种后端：`anthropic` / `openai` / `deepseek`，通过 `model_type` 配置切换。

---

## 6. Sub-Agent 工具权限

| Agent | 可用工具 |
|---|---|
| `novel_outline_agent` | 读大纲、写大纲 |
| `novel_write_agent` | 读大纲、列章节、读文件、写文件、编辑文件 |
| `novel_review_agent` | 列章节、读文件、读大纲（只读，无写入） |

---

## 7. 容错机制

- **LLM 429 重试**：最多 3 次，`isRetryAble` 判断是否可重试
- **工具错误隔离**：`SafeToolMiddleware` 将错误转为字符串，会话不中断
- **历史压缩**：防止上下文超限导致请求失败
- **未知工具 no-op**：模型调用不存在的工具时，返回提示让其重新选择

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
| ChatTemplate | `prompt.FromMessages(schema.FString, ...)` | `map[string]any` → `[]*schema.Message` | 将模板变量（genre/concept/rule_lines 等）渲染为 System + User 消息 |
| ChatModel | Lite LLM | `[]*schema.Message` → `*schema.Message` | 调用轻量模型，返回 `{"selected_ids": [...]}` JSON |
| Lambda | `parseSelectionResponse` | `*schema.Message` → `string` | 解析 JSON，查找 ID 对应规则，格式化为注入文本 |

### 降级策略

任何 Chain 内部失败（模型初始化失败、LLM 调用错误、JSON 格式无效、筛选结果为空）均触发降级，注入用户所有 enabled 规则。降级逻辑收敛在 `buildRulesContent` 一处，不分散。

### System Prompt 注入位置

主 Agent 系统提示末尾有 `{user_rules}` 占位符（`prompt.go`），`NewMainAgent` 在初始化时替换为筛选后的规则文本（无规则时替换为空字符串）。
