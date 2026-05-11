package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"novelflow/database/mongodb"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/summarization"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
)

type AgentRunner struct {
	*adk.Runner
	*Session
}

type AgentRunnerConfig struct {
	*deep.Config
	*mongodb.MongoClient
	SID          string
	SystemPrompt string
	Session      *Session
}

type StreamFunc func(Message) bool

func NewAgentRunner(ctx context.Context, config *AgentRunnerConfig) (*AgentRunner, error) {
	var err error = nil

	config.ChatModel, err = getChatModel(ctx)
	if err != nil {
		return nil, err
	}

	if config.ModelRetryConfig == nil {
		config.ModelRetryConfig = &adk.ModelRetryConfig{
			MaxRetries: 3,
			IsRetryAble: func(ctx context.Context, err error) bool {
				return strings.Contains(err.Error(), "429") ||
					strings.Contains(err.Error(), "Too Many Request")
			},
		}
	}

	// 设置 UnknownToolsHandler 处理未定义的工具调用，让模型能够继续运行
	config.ToolsConfig.UnknownToolsHandler = func(ctx context.Context, name, input string) (string, error) {
		return fmt.Sprintf("[tool error]: tool %s is not defined. Please use an available tool.", name), nil
	}

	// 添加中间件
	summarizationMW, err := buildSummarization(ctx)
	if err != nil {
		return nil, err
	}
	config.Handlers = append(config.Handlers, &safeToolMiddleware{}, summarizationMW)

	a, err := deep.New(ctx, config.Config)
	if err != nil {
		return nil, err
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a,
		EnableStreaming: true,
	})

	var s *Session
	if config.Session != nil {
		s = config.Session
	} else {
		s, err = NewSession(ctx, config.SID, config.MongoClient)
		if err != nil {
			return nil, err
		}
	}
	s.Append(Message{
		Type:    ContentType,
		Role:    SystemRole,
		Content: config.SystemPrompt,
	})

	return &AgentRunner{
		Runner:  r,
		Session: s,
	}, nil
}

func (ar *AgentRunner) RunA(ctx context.Context, message Message, handlerFunc StreamFunc, opts ...adk.AgentRunOption) error {
	ar.Session.Append(message)
	resp := ar.Run(ctx, ar.Session.Use(), opts...)
	for {
		e, flag := resp.Next()
		if !flag {
			break
		}
		if e.Err != nil {
			return e.Err
		}

		if e.Output == nil || e.Output.MessageOutput == nil {
			continue
		}

		sm := e.Output.MessageOutput.MessageStream
		output := ""
		if sm != nil {
			defer sm.Close()
			for {
				m, err := sm.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return err
				}
				if m.Content != "" {
					handlerFunc(Message{
						Type:    ContentType,
						Content: m.Content,
					})
					output += m.Content
				}
				if m.ReasoningContent != "" {
					handlerFunc(Message{
						Type:    ThinkingType,
						Content: m.ReasoningContent,
					})
				}
			}
		}
		if output != "" {
			ar.Session.Append(Message{
				Type:    ContentType,
				Role:    AgentRole,
				Content: output,
			})
		}
		tm := e.Output.MessageOutput.ToolName
		if tm != "" {
			handlerFunc(Message{
				Type:    ToolType,
				Content: tm,
			})
		}
	}
	return nil
}

func buildSummarization(ctx context.Context) (adk.ChatModelAgentMiddleware, error) {
	cm, err := getLiteChatModel(ctx)
	if err != nil {
		return nil, err
	}

	summarizationMW, err := summarization.New(ctx, &summarization.Config{
		Model: cm, // 用于生成摘要的模型
		Trigger: &summarization.TriggerCondition{
			ContextTokens: 200000, // 触发摘要的 token 阈值
		},
	})

	return summarizationMW, err
}
