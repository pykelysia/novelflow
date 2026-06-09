package agent

import (
	"context"
	"fmt"
	"novelflow/config"
	"novelflow/database/mongodb"
	"testing"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/spf13/viper"
)

func newTestModel(ctx context.Context) *claude.ChatModel {
	baseurl := viper.GetString("llm.base_url")
	cm, err := claude.NewChatModel(ctx, &claude.Config{
		Model:     viper.GetString("llm.model_name"),
		BaseURL:   &baseurl,
		APIKey:    viper.GetString("llm.api_key"),
		MaxTokens: 20000,
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

	r, err := NewAgent(ctx, &Config{
		Config: &deep.Config{
			Name:        "test agent",
			Description: "test is run able",
			ChatModel:   cm,
		},
		MongoClient: mdb,
		UserID:      0,
	})
	if err != nil {
		t.Error(err)
	}

	err = r.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "你好",
	}, func(m Message) bool {
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

	r, err := NewAgent(ctx, &Config{
		Config: &deep.Config{
			Name:        "test agent",
			Description: "test is run able",
			ChatModel:   cm,
		},
		MongoClient: mdb,
		UserID:      0,
	})
	if err != nil {
		t.Error(err)
	}

	err = r.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "这个文件夹下有什么内容",
	}, func(m Message) bool {
		fmt.Print(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("\nSession %s end\n", r.SessionID)
}

func TestRunnerSession(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")
	cm := newTestModel(ctx)
	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		t.Error(err)
	}

	r, err := NewAgent(ctx, &Config{
		Config: &deep.Config{
			Name:      "test agent",
			ChatModel: cm,
		},
		MongoClient: mdb,
		UserID:      0,
	})
	if err != nil {
		t.Error(err)
	}

	contentType := ""
	err = r.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "这个文件夹下有什么内容",
	}, func(m Message) bool {
		if contentType != m.Type {
			fmt.Printf("\n%v > ", m.Type)
			contentType = m.Type
		}

		fmt.Print(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("\nSession %s end\n", r.SessionID)
}

func TestRunnerSkills(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig("../config.yaml")
	cm := newTestModel(ctx)
	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		t.Error(err)
	}

	skillsBackend, _ := local.NewBackend(ctx, &local.Config{})
	skillsMiddlewareBackend, _ := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: skillsBackend,
		BaseDir: ".skills",
	})
	skillsMiddleware, _ := skill.NewMiddleware(ctx, &skill.Config{
		Backend: skillsMiddlewareBackend,
	})

	r, err := NewAgent(ctx, &Config{
		Config: &deep.Config{
			Name:      "test agent",
			ChatModel: cm,
			Handlers: []adk.ChatModelAgentMiddleware{
				skillsMiddleware,
			},
		},
		MongoClient: mdb,
		UserID:      0,
	})
	if err != nil {
		t.Error(err)
	}

	contentType := ""
	err = r.RunA(ctx, Message{
		Type:    ContentType,
		Role:    UserRole,
		Content: "告诉我，如何起名一个小说标题",
	}, func(m Message) bool {
		if contentType != m.Type {
			fmt.Printf("\n%v > ", m.Type)
			contentType = m.Type
		}

		fmt.Print(m.Content)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("\nSession %s end\n", r.SessionID)
}
