package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	agent "novelflow/agents"
	"novelflow/backend/internal/servicecontext"
	"novelflow/cache"
	"novelflow/database/mongodb"
	"novelflow/database/task"

	"github.com/google/uuid"
	"github.com/spf13/viper"
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

type ChapterResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type GenerateResultResponse struct {
	SessionID string          `json:"session_id"`
	Status    string          `json:"status"`
	Chapters  []*ChapterResult `json:"chapters,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type TaskItem struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListTasksResponse struct {
	Tasks []*TaskItem `json:"tasks"`
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
	go s.runGeneration(svc.MongoDB, svc.RedisClient, svc.Ctx, &svc.WG, sessionID, userID, req)

	return &GenerateResponse{
		SessionID: sessionID,
		Status:    string(task.TaskPending),
	}, nil
}

func (s *GenerateService) GetGenerationStatus(svc *servicecontext.ServiceContext, userID uint, sessionID string) (*GenerateStatusResponse, error) {
	t, err := getCachedTask(context.Background(), svc, sessionID)
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

func (s *GenerateService) GetGenerationResult(svc *servicecontext.ServiceContext, userID uint, sessionID string) (*GenerateResultResponse, error) {
	if strings.Contains(sessionID, "..") {
		return nil, ErrTaskNotFound
	}

	t, err := getCachedTask(context.Background(), svc, sessionID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTaskNotFound
	}
	if t.UserID != userID {
		return nil, ErrTaskForbidden
	}

	resp := &GenerateResultResponse{
		SessionID: t.SessionID,
		Status:    string(t.Status),
		Error:     t.Error,
	}

	if t.Status != task.TaskCompleted {
		return resp, nil
	}

	dir := filepath.Join(viper.GetString("storage.novels_dir"), sessionID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return resp, nil
		}
		return nil, fmt.Errorf("read novels dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		resp.Chapters = append(resp.Chapters, &ChapterResult{
			Title:   strings.TrimSuffix(e.Name(), ".txt"),
			Content: string(data),
		})
	}

	sort.Slice(resp.Chapters, func(i, j int) bool {
		return resp.Chapters[i].Title < resp.Chapters[j].Title
	})

	return resp, nil
}

func (s *GenerateService) ListUserTasks(svc *servicecontext.ServiceContext, userID uint) (*ListTasksResponse, error) {
	tasks, err := task.ListUserTasks(context.Background(), svc.MongoDB, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	items := make([]*TaskItem, 0, len(tasks))
	for i := range tasks {
		items = append(items, &TaskItem{
			SessionID: tasks[i].SessionID,
			Status:    string(tasks[i].Status),
			Error:     tasks[i].Error,
			CreatedAt: tasks[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: tasks[i].UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &ListTasksResponse{Tasks: items}, nil
}

// getCachedTask 先查 Redis，miss 时查 MongoDB 并回填缓存
func getCachedTask(ctx context.Context, svc *servicecontext.ServiceContext, sessionID string) (*task.Task, error) {
	key := cache.TaskKeyPrefix + sessionID

	var t task.Task
	if hit, err := svc.RedisClient.GetJSON(ctx, key, &t); err == nil && hit {
		return &t, nil
	}

	result, err := task.GetTask(ctx, svc.MongoDB, sessionID)
	if err != nil || result == nil {
		return result, err
	}

	_ = svc.RedisClient.SetJSON(ctx, key, result, cache.TaskCacheTTL)
	return result, nil
}

// invalidateTaskCache 使任务缓存失效
func invalidateTaskCache(ctx context.Context, rc *cache.Client, sessionID string) {
	_ = rc.Del(ctx, cache.TaskKeyPrefix+sessionID)
}

func (s *GenerateService) runGeneration(mdb *mongodb.MongoClient, rc *cache.Client, parentCtx context.Context, wg *sync.WaitGroup, sessionID string, userID uint, req *GenerateRequest) {
	defer wg.Done()
	defer func() { <-s.sem }()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	prompt, err := composePrompt(req)
	if err != nil {
		task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskFailed, err.Error())
		invalidateTaskCache(ctx, rc, sessionID)
		return
	}

	if err := task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskRunning, ""); err != nil {
		return
	}
	invalidateTaskCache(ctx, rc, sessionID)

	a, err := agent.NewMainAgent(ctx, sessionID, userID)
	if err != nil {
		task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskFailed, err.Error())
		invalidateTaskCache(ctx, rc, sessionID)
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
		invalidateTaskCache(ctx, rc, sessionID)
		return
	}

	task.UpdateTaskStatus(ctx, mdb, sessionID, task.TaskCompleted, "")
	invalidateTaskCache(ctx, rc, sessionID)
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
