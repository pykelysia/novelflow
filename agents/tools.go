package agent

import (
	"github.com/cloudwego/eino/components/tool"
)

func loadAgentTools(sessionID string) (tools []tool.BaseTool) {
	tools = append(tools, writeFileTool(sessionID))
	return
}
