package subagents

import (
	"context"

	"novelflow/agents/core"
	"novelflow/agents/tools"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// CreateReviewAgent creates a review agent that examines generated novel chapters
// against the writing skills requirements and produces a scored quality report.
// It has read-only access to chapter and outline files and no write/edit capabilities.
func CreateReviewAgent(ctx context.Context, sessionID string) (adk.Agent, error) {
	cm, err := core.GetChatModel(ctx)
	if err != nil {
		return nil, err
	}

	skillsMiddleware, err := core.GetSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}

	agentTools := []tool.BaseTool{
		tools.ListChapterFilesTool(sessionID),
		tools.ReadFileTool(sessionID),
		tools.ReadOutlineFileTool(sessionID),
	}

	return deep.New(ctx, &deep.Config{
		Name:        "novel_review_agent",
		Description: "审查小说章节内容质量，检查是否满足各项写作技能的专家",
		ChatModel:   cm,
		Instruction: reviewAgentSystemPrompt,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               agentTools,
				UnknownToolsHandler: core.DefaultUnknownToolHandler,
			},
		},
		WithoutGeneralSubAgent: true,
		WithoutWriteTodos:      true,
		Handlers: []adk.ChatModelAgentMiddleware{
			skillsMiddleware, &core.SafeToolMiddleware{},
		},
	})
}
