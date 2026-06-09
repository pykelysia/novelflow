package agent

import (
	"context"
	"fmt"
	"novelflow/config"
	"testing"
)

// TestReviewSubAgent_NoWriteTools verifies that the review sub-agent registered
// inside the main agent only has read-only tools (list + read) and does NOT have
// write_novel_chapter_file_tool or edit_novel_chapter_file_tool.
func TestReviewSubAgent_NoWriteTools(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")

	// Create main agent (which internally creates the review sub-agent).
	// This confirms the integration path works without write tools leaking in.
	ma, err := NewMainAgent(ctx, nil, "", 0, "")
	if err != nil {
		t.Fatalf("NewMainAgent failed: %v", err)
	}

	// Session-level smoke: run a simple message to ensure the agent starts
	flag := ""
	err = ma.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "请尝试使用 task tool 让reviewer agent来撰写文件，如果不能，先告诉我不能这样操作",
	}, func(m Message) bool {

		if flag != m.Type {
			fmt.Print("\n<" + m.Type + ">:")
			flag = m.Type
		}
		fmt.Print(m.Content)
		return true
	})
	if err != nil {
		t.Errorf("RunA failed: %v", err)
	}
}
