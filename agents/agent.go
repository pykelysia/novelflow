package agent

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
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	*adk.Runner
	*Session
}

type Config struct {
	*deep.Config
	*mongodb.MongoClient
	SID          string
	SystemPrompt string
	Session      *Session
}

type StreamFunc func(Message) bool

func NewAgent(ctx context.Context, config *Config) (*Agent, error) {
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
	config.Handlers = append(config.Handlers, &SafeToolMiddleware{}, summarizationMW)

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

	return &Agent{
		Runner:  r,
		Session: s,
	}, nil
}

func (a *Agent) RunA(ctx context.Context, message Message, handlerFunc StreamFunc, opts ...adk.AgentRunOption) error {
	a.Session.Append(message)
	resp := a.Run(ctx, a.Session.Use(), opts...)
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
			a.Session.Append(Message{
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

			saveMsg := Message{
				Type: ToolResultType,
				Role: e.Output.MessageOutput.Role,
			}
			role := e.Output.MessageOutput.Role
			if sm := e.Output.MessageOutput.Message; sm != nil {
				if role == schema.Tool && sm.Content != "" {
					saveMsg.ToolResult = sm.Content
				}
				if len(sm.ToolCalls) > 0 {
					saveMsg.Content = tm + "\n" + sm.ToolCalls[0].Function.Arguments
				}
			}
			if saveMsg.Content == "" {
				saveMsg.Content = tm
			}
			a.Session.Append(saveMsg)
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
		Model: cm,
		Trigger: &summarization.TriggerCondition{
			ContextTokens: 200000,
		},
	})

	return summarizationMW, err
}
