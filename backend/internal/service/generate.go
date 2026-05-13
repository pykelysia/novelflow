package service

import (
	"context"
	"fmt"
	"strings"

	agent "novelflow/agents"
	"novelflow/backend/internal/servicecontext"
	"novelflow/database/mongodb"
	"novelflow/database/task"

	"github.com/google/uuid"
)

type GenerateService struct{}

func NewGenerateService() *GenerateService {
	return &GenerateService{}
}

type GenerateRequest struct {
	Genre        string `json:"genre" binding:"required"`
	Concept      string `json:"concept" binding:"required"`
	Protagonist  string `json:"protagonist,omitempty"`
	WorldSetting string `json:"world_setting,omitempty"`
	ChapterCount int    `json:"chapter_count"`
	Style        string `json:"style,omitempty"`
	Requirements string `json:"requirements,omitempty"`
}

type GenerateResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

type GenerateStatusResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *GenerateService) StartGeneration(svc *servicecontext.ServiceContext, userID uint, req *GenerateRequest) (*GenerateResponse, error) {
	sessionID := uuid.New().String()

	if req.ChapterCount <= 0 {
		req.ChapterCount = 1
	}

	_, err := task.CreateTask(context.Background(), svc.MongoDB, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go runGeneration(ctx, cancel, svc.MongoDB, sessionID, userID, req)

	return &GenerateResponse{
		SessionID: sessionID,
		Status:    string(task.TaskPending),
	}, nil
}

func (s *GenerateService) GetGenerationStatus(svc *servicecontext.ServiceContext, userID uint, sessionID string) (*GenerateStatusResponse, error) {
	t, err := task.GetTask(context.Background(), svc.MongoDB, sessionID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTaskNotFound
	}
	if t.UserID != userID {
		return nil, ErrTaskForbidden
	}

	return &GenerateStatusResponse{
		SessionID: t.SessionID,
		Status:    string(t.Status),
		Error:     t.Error,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func runGeneration(ctx context.Context, cancel context.CancelFunc, mdb *mongodb.MongoClient, sessionID string, userID uint, req *GenerateRequest) {
	defer cancel()
	prompt := composePrompt(req)

	if err := task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskRunning, ""); err != nil {
		return
	}

	a, err := agent.NewMainAgent(ctx, sessionID, userID)
	if err != nil {
		task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskFailed, err.Error())
		return
	}

	err = a.RunA(ctx, agent.Message{
		Type:    agent.ContentType,
		Role:    agent.UserRole,
		Content: prompt,
	}, func(msg agent.Message) bool {
		return true
	})
	if err != nil {
		task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskFailed, err.Error())
		return
	}

	task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskCompleted, "")
}

func composePrompt(req *GenerateRequest) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("请创作一部%s小说。\n\n", req.Genre))
	b.WriteString(req.Concept)
	b.WriteString("\n")

	if req.Protagonist != "" {
		b.WriteString(fmt.Sprintf("\n主角设定：%s", req.Protagonist))
	}
	if req.WorldSetting != "" {
		b.WriteString(fmt.Sprintf("\n世界观设定：%s", req.WorldSetting))
	}
	if req.ChapterCount > 0 {
		b.WriteString(fmt.Sprintf("\n请生成 %d 章内容。", req.ChapterCount))
	}
	if req.Style != "" {
		b.WriteString(fmt.Sprintf("\n风格要求：%s", req.Style))
	}
	if req.Requirements != "" {
		b.WriteString(fmt.Sprintf("\n其他要求：%s", req.Requirements))
	}

	return b.String()
}
