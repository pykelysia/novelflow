package agents

import (
	"context"
	"fmt"
	"novelflow/config"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

func TestRunner(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")

	baseurl := viper.GetString("llm.base_url")
	cm, err := claude.NewChatModel(ctx, &claude.Config{
		Model:   viper.GetString("llm.model_name"),
		BaseURL: &baseurl,
		APIKey:  viper.GetString("llm.api_key"),
	})

	r, err := NewAgentRunner(ctx, &AgentRunnerConfig{
		Config: &deep.Config{
			Name:        "test agent",
			Description: "test is run able",
			ChatModel:   cm,
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = r.RunA(ctx, []adk.Message{schema.UserMessage("这个文件夹下有哪些文件")}, func(m Message) bool {
		fmt.Println(m.Content)
		return true
	})
	if err != nil {
		fmt.Println(err)
	}
}
