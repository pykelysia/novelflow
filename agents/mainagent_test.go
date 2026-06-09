package agent

import (
	"context"
	"fmt"
	"novelflow/config"
	"testing"
)

func TestMainAgent(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")

	ma, err := NewMainAgent(ctx, nil, "", 0, "")
	if err != nil {
		t.Error(err)
	}

	flag := ""
	err = ma.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "请帮我写一个短篇小说，题材是科幻，内容要有趣，字数在10000字左右。要求适当分章节。",
	}, func(m Message) bool {
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
