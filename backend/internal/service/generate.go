package service

import (
	"context"
	"fmt"

	agent "novelflow/agents"
	"novelflow/backend/internal/servicecontext"
	"novelflow/database/task"

	"github.com/google/uuid"
)

type GenerateService struct{}

func NewGenerateService() *GenerateService {
	return &GenerateService{}
}

type GenerateRequest struct {
	Prompt string `json:"prompt" binding:"required"`
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

	_, err := task.CreateTask(context.Background(), svc.MongoDB, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	go runGeneration(svc, sessionID, userID, req.Prompt)

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

func runGeneration(svc *servicecontext.ServiceContext, sessionID string, userID uint, prompt string) {
	ctx := context.Background()

	if err := task.UpdateTaskStatus(ctx, svc.MongoDB, sessionID, task.TaskRunning, ""); err != nil {
		return
	}

	a, err := agent.NewMainAgent(ctx, sessionID, userID)
	if err != nil {
		task.UpdateTaskStatus(ctx, svc.MongoDB, sessionID, task.TaskFailed, err.Error())
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
		task.UpdateTaskStatus(ctx, svc.MongoDB, sessionID, task.TaskFailed, err.Error())
		return
	}

	task.UpdateTaskStatus(ctx, svc.MongoDB, sessionID, task.TaskCompleted, "")
}
