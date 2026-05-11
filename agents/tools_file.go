package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

func writeFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		writeFileToolInfo(),
		writeFileToolInvoke(sessionID),
	)
}

type writeFileToolInput struct {
	Titile  string
	Content string
}

func writeFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "write_novel_chapter_file_tool",
		Desc: "create one file, and write a single chapter to the file",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"title": {
					Type: "string",
					Desc: "the title of chapter, which will be used as file name",
				},
				"content": {
					Type: "string",
					Desc: "the whole content to write in file",
				},
			},
		),
	}
}

func writeFileToolInvoke(sessionID string) utils.InvokeFunc[writeFileToolInput, string] {
	return func(ctx context.Context, input writeFileToolInput) (output string, err error) {
		dir := filepath.Join(".", "data", "novels", sessionID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}

		path := filepath.Join(dir, input.Titile+".txt")
		file, err := os.Create(path)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Write([]byte(input.Content))
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("[tool result]成功写入 %s", path), nil
	}
}

func readFileTool() tool.BaseTool {
	return utils.NewTool(
		readFileToolInfo(),
		readFileToolInvoke(),
	)
}

type readFileToolInput struct{}

func readFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{}
}

func readFileToolInvoke() utils.InvokeFunc[readFileToolInput, string] {
	return func(ctx context.Context, input readFileToolInput) (output string, err error) {
		return
	}
}
