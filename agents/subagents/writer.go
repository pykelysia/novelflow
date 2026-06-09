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

// CreateWriteAgent creates a write agent that generates novel chapter content.
// It reads the outline and existing chapters for context, then writes new chapter
// files following the outline's chapter plan and writing skill requirements.
func CreateWriteAgent(ctx context.Context, sessionID string) (adk.Agent, error) {
	cm, err := core.GetChatModel(ctx)
	if err != nil {
		return nil, err
	}

	skillsMiddleware, err := core.GetSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}

	agentTools := []tool.BaseTool{
		tools.WriteFileTool(sessionID),
		tools.EditFileTool(sessionID),
		tools.ReadOutlineFileTool(sessionID),
		tools.ListChapterFilesTool(sessionID),
		tools.ReadFileTool(sessionID),
	}

	return deep.New(ctx, &deep.Config{
		Name:        "novel_write_agent",
		Description: "撰写小说章节的专家，根据大纲创作各章节内容",
		ChatModel:   cm,
		Instruction: writeAgentSystemPrompt,
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
