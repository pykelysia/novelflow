package agents

import (
	"context"
	"fmt"
	"novelflow/config"
	"novelflow/database/mongodb"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

func newTestModel(ctx context.Context) *claude.ChatModel {
	baseurl := viper.GetString("llm.base_url")
	cm, err := claude.NewChatModel(ctx, &claude.Config{
		Model:   viper.GetString("llm.model_name"),
		BaseURL: &baseurl,
		APIKey:  viper.GetString("llm.api_key"),
	})
	if err != nil {
		panic(err)
	}
	return cm
}

func TestRunner(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")
	cm := newTestModel(ctx)

	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		t.Error(err)
	}

	r, err := NewAgentRunner(ctx, &AgentRunnerConfig{
		Config: &deep.Config{
			Name:        "test agent",
			Description: "test is run able",
			ChatModel:   cm,
		},
		MongoClient: mdb,
	})
	if err != nil {
		t.Error(err)
	}

	err = r.RunA(ctx, []adk.Message{schema.UserMessage("你好")}, func(m Message) bool {
		fmt.Println(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Println("end")
}

// 测试调用一个不存在的工具，验证 UnknownToolsHandler 是否生效。
func TestRunnerWithNoTool(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")
	cm := newTestModel(ctx)
	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		t.Error(err)
	}

	r, err := NewAgentRunner(ctx, &AgentRunnerConfig{
		Config: &deep.Config{
			Name:        "test agent",
			Description: "test is run able",
			ChatModel:   cm,
		},
		MongoClient: mdb,
	})
	if err != nil {
		t.Error(err)
	}

	err = r.RunA(ctx, []adk.Message{schema.UserMessage("这个文件夹下有什么内容")}, func(m Message) bool {
		fmt.Println(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Println("end")
}
