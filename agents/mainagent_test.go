package agents

import (
	"context"
	"fmt"
	"novelflow/agents/runner"
	"novelflow/config"
	"testing"
)

func TestMainAgent(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")

	ma, err := NewMainAgent(ctx, "")
	if err != nil {
		t.Error(err)
	}

	flag := ""
	err = ma.RunA(ctx, runner.Message{
		Type:    runner.ContentType,
		Role:    runner.UserRole,
		Content: "请帮我写一个短篇小说，题材是科幻，内容要有趣，字数在1000字左右。",
	}, func(m runner.Message) bool {
		if flag != m.Type {
			fmt.Print("\n<" + m.Type + ">:")
			flag = m.Type
		}
		fmt.Print(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
}
