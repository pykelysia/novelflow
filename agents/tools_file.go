package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

func writeFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		writeFileToolInfo(),
		writeFileToolInvoke(sessionID),
	)
}

type writeFileToolInput struct {
	Title   string
	Content string
}

func writeFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "write_novel_chapter_file_tool",
		Desc: "create one file, and write a single chapter to the file",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"title": {
					Type:     "string",
					Desc:     "the title of chapter, which will be used as file name",
					Required: true,
				},
				"content": {
					Type:     "string",
					Desc:     "the whole content to write in file",
					Required: true,
				},
			},
		),
	}
}

func writeFileToolInvoke(sessionID string) utils.InvokeFunc[writeFileToolInput, string] {
	return func(ctx context.Context, input writeFileToolInput) (output string, err error) {
		dir := filepath.Join(viper.GetString("storage.novels_dir"), sessionID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}

		path := filepath.Join(dir, input.Title+".txt")
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

func listChapterFilesTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		listChapterFilesToolInfo(),
		listChapterFilesToolInvoke(sessionID),
	)
}

func listChapterFilesToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "list_chapter_files_tool",
		Desc: "list all chapter files in the session directory, returns the list of chapter titles",
	}
}

func listChapterFilesToolInvoke(sessionID string) utils.InvokeFunc[struct{}, string] {
	return func(ctx context.Context, _ struct{}) (output string, err error) {
		if sessionID == "" {
			return "", fmt.Errorf("sessionID is required")
		}
		dir := filepath.Join(viper.GetString("storage.novels_dir"), sessionID)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return "当前没有任何已生成的章节文件。", nil
			}
			return "", fmt.Errorf("failed to read directory: %v", err)
		}

		var titles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
				titles = append(titles, strings.TrimSuffix(e.Name(), ".txt"))
			}
		}
		if len(titles) == 0 {
			return "当前没有任何已生成的章节文件。", nil
		}
		return "已生成的章节列表：\n" + strings.Join(titles, "\n"), nil
	}
}

func readFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		readFileToolInfo(),
		readFileToolInvoke(sessionID),
	)
}

type readFileToolInput struct {
	Title string `json:"title"`
}

func readFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "read_novel_chapter_file_tool",
		Desc: "read a novel chapter file by title, return the full content of the file",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"title": {
					Type:     "string",
					Desc:     "the title of the chapter to read, which matches the filename (without .txt extension)",
					Required: true,
				},
			},
		),
	}
}

func readFileToolInvoke(sessionID string) utils.InvokeFunc[readFileToolInput, string] {
	return func(ctx context.Context, input readFileToolInput) (output string, err error) {
		if strings.Contains(input.Title, "..") {
			return "", fmt.Errorf("invalid title: '..' is not allowed")
		}

		path := filepath.Join(viper.GetString("storage.novels_dir"), sessionID, input.Title+".txt")

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("chapter file not found: %s", input.Title+".txt")
			}
			return "", fmt.Errorf("failed to read chapter file %s: %v", input.Title+".txt", err)
		}

		return string(data), nil
	}
}
