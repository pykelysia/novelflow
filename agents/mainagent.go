package agent

import (
	"context"
	"fmt"
	"strings"

	"novelflow/agents/core"
	"novelflow/agents/session"
	"novelflow/agents/subagents"
	"novelflow/agents/tools"
	"novelflow/database/mongodb"
	"novelflow/database/mysql"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func NewMainAgent(ctx context.Context, mdb *mongodb.MongoClient, sessionID string, userID uint, rulesContent string) (*Agent, error) {
	var err error
	if mdb == nil {
		mdb, err = mongodb.NewMongoDB()
		if err != nil {
			return nil, err
		}
	}

	sess, err := session.NewSession(ctx, sessionID, userID, mdb)
	if err != nil {
		return nil, err
	}
	resolvedID := sess.SessionPart.SID

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

	cfg := &core.Config{
		Config: &deep.Config{
			Name:        "novelflow agent",
			Description: "an agent to write novel, you can ask it to generate a short novel.",
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						tools.ReadFileTool(resolvedID),
					},
				},
			},
		},
		MongoClient:  mdb,
		Session:      sess,
		SystemPrompt: strings.ReplaceAll(
				strings.ReplaceAll(mainAgentSystemPrompt, "{session_id}", resolvedID),
				"{user_rules}", rulesContent,
			),
	}

	// 创建质量审查 sub-agent
	reviewAgent, err := subagents.CreateReviewAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create review sub-agent: %v", err)
	}

	// 创建大纲 sub-agent
	outlineAgent, err := subagents.CreateOutlineAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create outline sub-agent: %v", err)
	}

	// 创建写作 sub-agent
	writeAgent, err := subagents.CreateWriteAgent(ctx, resolvedID)
	if err != nil {
		return nil, fmt.Errorf("failed to create write sub-agent: %v", err)
	}

	cfg.Config.SubAgents = []adk.Agent{outlineAgent, writeAgent, reviewAgent}

	skillsSystem, err := core.GetSkillsSystem(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Handlers = append(cfg.Handlers, skillsSystem)

	r, err := core.NewAgent(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return r, nil
}
