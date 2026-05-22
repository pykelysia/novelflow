package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// CreateOutlineAgent creates an outline agent that generates and modifies
// the novel's outline/章纲. It has access to volume-outline and related
// skill modules, and can read/write outline.md files.
func CreateOutlineAgent(ctx context.Context, sessionID string) (adk.Agent, error) {
	cm, err := getChatModel(ctx)
	if err != nil {
		return nil, err
	}

	skillsMiddleware, err := getSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}

	tools := []tool.BaseTool{
		writeOutlineFileTool(sessionID),
		readOutlineFileTool(sessionID),
	}

	return deep.New(ctx, &deep.Config{
		Name:        "novel_outline_agent",
		Description: "生成和修改小说大纲的专家，负责分卷章纲规划",
		ChatModel:   cm,
		Instruction: outlineAgentSystemPrompt,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("[tool error]: tool %s is not defined. Please use write_outline_file_tool or read_outline_file_tool.", name), nil
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
