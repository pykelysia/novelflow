package service

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"

	agent "novelflow/agents"
	"novelflow/backend/internal/servicecontext"
	"novelflow/database/mongodb"
	"novelflow/database/task"

	"github.com/google/uuid"
)

const MaxConcurrentGenerations = 5

type GenerateService struct {
	sem chan struct{}
}

func NewGenerateService() *GenerateService {
	return &GenerateService{
		sem: make(chan struct{}, MaxConcurrentGenerations),
	}
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

	select {
	case s.sem <- struct{}{}:
	default:
		return nil, ErrTooManyRequests
	}

	_, err := task.CreateTask(context.Background(), svc.MongoDB, sessionID, userID)
	if err != nil {
		<-s.sem
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	svc.WG.Add(1)
	go s.runGeneration(svc.MongoDB, svc.Ctx, &svc.WG, sessionID, userID, req)

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

func (s *GenerateService) runGeneration(mdb *mongodb.MongoClient, parentCtx context.Context, wg *sync.WaitGroup, sessionID string, userID uint, req *GenerateRequest) {
	defer wg.Done()
	defer func() { <-s.sem }()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	prompt, err := composePrompt(req)
	if err != nil {
		task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskFailed, err.Error())
		return
	}

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

var promptTmpl = template.Must(template.New("prompt").Parse(`请创作一部{{.Genre}}小说。

{{.Concept}}
{{if .Protagonist}}
主角设定：{{.Protagonist}}{{end}}{{if .WorldSetting}}
世界观设定：{{.WorldSetting}}{{end}}{{if gt .ChapterCount 0}}
请生成 {{.ChapterCount}} 章内容。{{end}}{{if .Style}}
风格要求：{{.Style}}{{end}}{{if .Requirements}}
其他要求：{{.Requirements}}{{end}}`))

func composePrompt(req *GenerateRequest) (string, error) {
	var buf bytes.Buffer
	if err := promptTmpl.Execute(&buf, req); err != nil {
		return "", fmt.Errorf("compose prompt: %w", err)
	}
	return buf.String(), nil
}
