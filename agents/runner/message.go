package runner

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

type Message struct {
	Type       string          `bson:"type"`
	Role       schema.RoleType `bson:"role"`
	Content    string          `bson:"content"`
	ToolResult string          `bson:"tool_result,omitempty"`
	SessionID  string          `bson:"session_id"`
	CreatedAt  time.Time       `bson:"created_at"`
}

const (
	UserRole  = schema.User
	AgentRole = schema.Assistant

	ContentType  = "Content"
	ThinkingType = "Thinking"
	ToolType     = "Tool"
)
