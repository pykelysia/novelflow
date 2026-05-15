package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// CreateReviewAgent creates a review agent that examines generated novel chapters
// against the writing skills requirements and produces a scored quality report.
// It has read-only access to chapter files and no write/edit capabilities.
func CreateReviewAgent(ctx context.Context, sessionID string, outline string) (adk.Agent, error) {
	cm, err := getChatModel(ctx)
	if err != nil {
		return nil, err
	}

	skillsMiddleware, err := getSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}

	prompt := strings.ReplaceAll(reviewAgentSystemPrompt, "{outline}", outline)

	tools := []tool.BaseTool{
		listChapterFilesTool(sessionID),
		readFileTool(sessionID),
	}

	return deep.New(ctx, &deep.Config{
		Name:        "novel_review_agent",
		Description: "审查小说章节内容质量，检查是否满足各项写作技能的专家",
		ChatModel:   cm,
		Instruction: prompt,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					if strings.Contains(name, "write") || strings.Contains(name, "edit") || strings.Contains(name, "delete") || strings.Contains(name, "remove") {
						return "[tool error]: 审查 agent 不允许执行写/删/改操作。请使用 list_chapter_files_tool 或 read_novel_chapter_file_tool。", nil
					}
					return fmt.Sprintf("[tool error]: tool %s is not defined. Please use an available tool.", name), nil
				},
			},
		},
		WithoutGeneralSubAgent: true,
		WithoutWriteTodos:      true,
		Handlers: []adk.ChatModelAgentMiddleware{
			skillsMiddleware, &SafeToolMiddleware{},
		},
	})
}
