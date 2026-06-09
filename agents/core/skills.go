package core

import (
	"context"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/spf13/viper"
)

func GetSkillsSystem(ctx context.Context) (adk.ChatModelAgentMiddleware, error) {
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
