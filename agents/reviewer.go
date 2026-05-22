package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// CreateReviewAgent creates a review agent that examines generated novel chapters
// against the writing skills requirements and produces a scored quality report.
// It has read-only access to chapter and outline files and no write/edit capabilities.
func CreateReviewAgent(ctx context.Context, sessionID string) (adk.Agent, error) {
	cm, err := getChatModel(ctx)
	if err != nil {
		return nil, err
	}

	skillsMiddleware, err := getSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}

	tools := []tool.BaseTool{
		listChapterFilesTool(sessionID),
		readFileTool(sessionID),
		readOutlineFileTool(sessionID),
	}

	return deep.New(ctx, &deep.Config{
		Name:        "novel_review_agent",
		Description: "审查小说章节内容质量，检查是否满足各项写作技能的专家",
		ChatModel:   cm,
		Instruction: reviewAgentSystemPrompt,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				UnknownToolsHandler: defaultUnknownToolHandler,
			},
		},
		WithoutGeneralSubAgent: true,
		WithoutWriteTodos:      true,
		Handlers: []adk.ChatModelAgentMiddleware{
			skillsMiddleware, &SafeToolMiddleware{},
		},
	})
}

