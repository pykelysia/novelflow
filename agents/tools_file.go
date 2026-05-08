package agents

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

func writeFileTool() tool.BaseTool {
	return utils.NewTool(
		writeFileToolInfo(),
		writeFileToolInvoke(),
	)
}

type writeFileToolInput struct {
	Path    string
	Content string
}

func writeFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "write_novel_chapter_file_tool",
		Desc: "create one file, and write a single chapter to the file",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"path": {
					Type: "string",
					Desc: "the path to file which need to write.",
				},
				"content": {
					Type: "string",
					Desc: "the whole content to write in file",
				},
			},
		),
	}
}

func writeFileToolInvoke() utils.InvokeFunc[writeFileToolInput, string] {
	return func(ctx context.Context, input writeFileToolInput) (output string, err error) {
		dir := filepath.Dir(input.Path)
		info, err := os.Stat(dir)
		if err != nil {
			err := os.Mkdir(dir, 0666)
			if err != nil {
				return "", err
			}
		}
		info, _ = os.Stat(dir)
		if flag := info.IsDir(); !flag {
			return "", err
		}

		file, err := os.Create(input.Path)
		if err != nil {
			return "", err
		}

		_, err = file.Write([]byte(input.Content))
		if err != nil {
			return "", err
		}

		return "[tool result]成功写入", nil
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
