package agent

import (
	"context"
	"fmt"
	"novelflow/database/mongodb"
	"novelflow/database/mysql"
	"strings"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/spf13/viper"
)

func NewMainAgent(ctx context.Context, sessionID string, userID uint) (*Agent, error) {
	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		return nil, err
	}

	session, err := NewSession(ctx, sessionID, userID, mdb)
	if err != nil {
		return nil, err
	}
	resolvedID := session.SessionPart.SID

	// 将用户-会话关联写入 MySQL（仅当 userID > 0）
	if userID > 0 {
		sqlDB, sqlErr := mysql.NewDB()
		if sqlErr == nil {
			userSessionRepo := mysql.NewUserSessionRepository(sqlDB)
			// TODO: 添加日志记录，便于排查 mapping 写入失败原因
			_ = userSessionRepo.Create(&mysql.UserSession{
				UserID:    userID,
				SessionID: resolvedID,
			})
		}
	}

	cfg := &Config{
		Config: &deep.Config{
			Name:        "novelflow agent",
			Description: "an agent to write novel, you can ask it to generate a short novel.",
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						readFileTool(sessionID),
					},
				},
			},
		},
		MongoClient:  mdb,
		Session:      session,
		SystemPrompt: strings.ReplaceAll(mainAgentSystemPrompt, "{session_id}", resolvedID),
	}

	// 创建质量审查 sub-agent
	reviewAgent, err := CreateReviewAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create review sub-agent: %v", err)
	}

	// 创建大纲 sub-agent
	outlineAgent, err := CreateOutlineAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create outline sub-agent: %v", err)
	}

	// 创建写作 sub-agent
	writeAgent, err := CreateWriteAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create write sub-agent: %v", err)
	}

	cfg.Config.SubAgents = []adk.Agent{outlineAgent, writeAgent, reviewAgent}

	skillsSystem, err := getSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Handlers = append(cfg.Handlers, skillsSystem)

	r, err := NewAgent(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return r, nil
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
