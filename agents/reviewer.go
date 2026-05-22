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
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					if isWriteToolName(name) {
						return "[tool error]: 审查 agent 不允许执行写/删/改操作。请使用 list_chapter_files_tool、read_novel_chapter_file_tool 或 read_outline_file_tool。", nil
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

// isWriteToolName checks whether a tool name corresponds to a write/delete/modify operation.
// The review sub-agent's UnknownToolsHandler uses this to block write-capable tools,
// ensuring the reviewer only has read-only access.
func isWriteToolName(name string) bool {
	return strings.Contains(name, "write") ||
		strings.Contains(name, "edit") ||
		strings.Contains(name, "delete") ||
		strings.Contains(name, "remove")
}
