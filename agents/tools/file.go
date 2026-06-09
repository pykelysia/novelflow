package tools

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

func WriteFileTool(sessionID string) tool.BaseTool {
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

func ListChapterFilesTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		listChapterFilesToolInfo(),
		listChapterFilesToolInvoke(sessionID),
	)
}

func listChapterFilesToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "list_chapter_files_tool",
		Desc: "list all chapter files in the session directory, returns the list of chapter titles",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
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

func ReadFileTool(sessionID string) tool.BaseTool {
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

type editFileToolInput struct {
	Title      string
	OldContent string
	NewContent string
}

func editFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "edit_novel_chapter_file_tool",
		Desc: "edit a novel chapter file by replacing specified old content with new content. Use this to modify or extend existing chapters without rewriting the entire file",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"title": {
					Type:     "string",
					Desc:     "the title of the chapter to edit, which matches the filename (without .txt extension)",
					Required: true,
				},
				"old_content": {
					Type:     "string",
					Desc:     "the exact existing text to be replaced",
					Required: true,
				},
				"new_content": {
					Type:     "string",
					Desc:     "the new text to replace the old content with",
					Required: true,
				},
			},
		),
	}
}

func editFileToolInvoke(sessionID string) utils.InvokeFunc[editFileToolInput, string] {
	return func(ctx context.Context, input editFileToolInput) (output string, err error) {
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

		content := string(data)
		if !strings.Contains(content, input.OldContent) {
			return "", fmt.Errorf("old_content not found in chapter file: %s", input.Title+".txt")
		}

		newContent := strings.Replace(content, input.OldContent, input.NewContent, 1)
		if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
			return "", fmt.Errorf("failed to write chapter file %s: %v", input.Title+".txt", err)
		}

		return fmt.Sprintf("[tool result]成功修改 %s", path), nil
	}
}

func EditFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		editFileToolInfo(),
		editFileToolInvoke(sessionID),
	)
}

func WriteOutlineFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		writeOutlineFileToolInfo(),
		writeOutlineFileToolInvoke(sessionID),
	)
}

type writeOutlineFileToolInput struct {
	Content string `json:"content"`
}

func writeOutlineFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "write_outline_file_tool",
		Desc: "write the novel outline to outline.md file. Use this to create or update the novel's outline/章纲.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"content": {
					Type:     "string",
					Desc:     "the outline content in markdown format",
					Required: true,
				},
			},
		),
	}
}

func writeOutlineFileToolInvoke(sessionID string) utils.InvokeFunc[writeOutlineFileToolInput, string] {
	return func(ctx context.Context, input writeOutlineFileToolInput) (output string, err error) {
		dir := filepath.Join(viper.GetString("storage.novels_dir"), sessionID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}

		path := filepath.Join(dir, "outline.md")
		if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
			return "", fmt.Errorf("failed to write outline: %v", err)
		}

		return fmt.Sprintf("[tool result]成功写入大纲文件 %s", path), nil
	}
}

func ReadOutlineFileTool(sessionID string) tool.BaseTool {
	return utils.NewTool(
		readOutlineFileToolInfo(),
		readOutlineFileToolInvoke(sessionID),
	)
}

func readOutlineFileToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "read_outline_file_tool",
		Desc: "read the current outline from outline.md file. Returns the full outline content.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}
}

func readOutlineFileToolInvoke(sessionID string) utils.InvokeFunc[struct{}, string] {
	return func(ctx context.Context, _ struct{}) (output string, err error) {
		if sessionID == "" {
			return "", fmt.Errorf("sessionID is required")
		}
		path := filepath.Join(viper.GetString("storage.novels_dir"), sessionID, "outline.md")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "当前还没有大纲文件。请先使用 write_outline_file_tool 创建大纲。", nil
			}
			return "", fmt.Errorf("failed to read outline: %v", err)
		}
		return string(data), nil
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
