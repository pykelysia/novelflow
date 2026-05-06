package agents

import (
	"context"
	"novelflow/agents/runner"
	"novelflow/database/mongodb"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/spf13/viper"
)

type InternalAgent struct {
	*runner.AgentRunner
}

func NewMainAgent(ctx context.Context, sessionID string) (*InternalAgent, error) {
	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		return nil, err
	}
	cfg := &runner.AgentRunnerConfig{
		Config: &deep.Config{
			Name:        "novelflow agent",
			Description: "an agent to write novel, you can ask it to generate a short novel.",
		},
		MongoClient:  mdb,
		SID:          sessionID,
		SystemPrompt: mainAgentSystemPrompt,
	}

	skillsSystem, err := getSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Handlers = append(cfg.Handlers, skillsSystem)

	r, err := runner.NewAgentRunner(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &InternalAgent{
		AgentRunner: r,
	}, nil
}

func getSkillsSystem(ctx context.Context) (adk.ChatModelAgentMiddleware, error) {
	skillDir := viper.GetString("skills.base_dir")
	if skillDir == "" {
		return nil, nil
	}

	backend, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		return nil, err
	}
	skillsBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: backend,
		BaseDir: skillDir,
	})
	if err != nil {
		return nil, err
	}
	skillsMiddleware, err := skill.NewMiddleware(ctx, &skill.Config{
		Backend: skillsBackend,
	})
	if err != nil {
		return nil, err
	}
	return skillsMiddleware, nil
}
