package agent

import (
	"context"

	"novelflow/agents/core"
	"novelflow/agents/session"
)

// 类型别名：外部代码继续用 agent.Message、agent.Agent 等，无需修改任何调用方。
type Message    = session.Message
type Agent      = core.Agent
type Config     = core.Config
type StreamFunc = core.StreamFunc

// 常量重导出：外部代码继续用 agent.ContentType、agent.UserRole 等。
const (
	ContentType    = session.ContentType
	ThinkingType   = session.ThinkingType
	ToolType       = session.ToolType
	ToolResultType = session.ToolResultType
	UserRole       = session.UserRole
	AgentRole      = session.AgentRole
	SystemRole     = session.SystemRole
)

// NewAgent 是 core.NewAgent 的包级入口，保持测试和外部代码的直接调用路径。
func NewAgent(ctx context.Context, config *Config) (*Agent, error) {
	return core.NewAgent(ctx, config)
}
