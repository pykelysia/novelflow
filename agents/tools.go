package agents

import (
	"github.com/cloudwego/eino/components/tool"
)

func loadAgentTools() (tools []tool.BaseTool) {
	tools = append(tools, writeFileTool())
	// tools = append(tools, re)
	return
}
