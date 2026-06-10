package tools

import (
	"context"
	"fmt"

	"novelflow/agents/session"
	"novelflow/database/mongodb"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type retrieveArchivedResultInput struct {
	ArchiveID string `json:"archive_id"`
}

func RetrieveArchivedResultTool(mdb *mongodb.MongoClient) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "retrieve_archived_tool_result",
			Desc: "根据归档ID查看被压缩归档的工具调用结果。当上下文压缩后，较早的工具调用结果会被归档，使用此工具可以查看完整结果。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"archive_id": {
					Type:     "string",
					Desc:     "工具调用归档ID，来自上下文摘要中的 [archive_id]",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input retrieveArchivedResultInput) (string, error) {
			arch, err := session.LoadArchive(ctx, mdb, input.ArchiveID)
			if err != nil {
				return "", fmt.Errorf("无法找到归档结果: %w", err)
			}
			status := "成功"
			if !arch.Success {
				status = "失败"
			}
			return fmt.Sprintf("[归档工具结果]\n工具: %s\n参数: %s\n状态: %s\n结果:\n%s",
				arch.ToolName, arch.ToolArgs, status, arch.Result), nil
		},
	)
}
