package agent

import (
	"context"
	"errors"
	"io"
	"novelflow/database/mongodb"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/summarization"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	*adk.Runner
	SessionID string
}

type Config struct {
	*deep.Config
	*mongodb.MongoClient
	SID          string
	SystemPrompt string
	Session      *Session
	UserID       uint
}

type StreamFunc func(Message) bool

func NewAgent(ctx context.Context, config *Config) (*Agent, error) {
	var err error

	config.ChatModel, err = getChatModel(ctx)
	if err != nil {
		return nil, err
	}

	if config.ModelRetryConfig == nil {
		config.ModelRetryConfig = &adk.ModelRetryConfig{
			MaxRetries:  3,
			IsRetryAble: isRetryAble,
		}
	}

	config.ToolsConfig.UnknownToolsHandler = defaultUnknownToolHandler

	var s *Session
	if config.Session != nil {
		s = config.Session
	} else {
		s, err = NewSession(ctx, config.SID, config.UserID, config.MongoClient)
		if err != nil {
			return nil, err
		}
	}

	// Promote SystemPrompt to the agent's Instruction so it is prepended by
	// genModelInput each turn rather than stored once in message history.
	if config.SystemPrompt != "" && config.Instruction == "" {
		config.Instruction = config.SystemPrompt
	}

	summarizationMW, err := buildSummarization(ctx)
	if err != nil {
		return nil, err
	}
	config.Handlers = append(config.Handlers, &SafeToolMiddleware{}, summarizationMW, newSessionMiddleware(s))

	a, err := deep.New(ctx, config.Config)
	if err != nil {
		return nil, err
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a,
		EnableStreaming: true,
	})

	return &Agent{Runner: r, SessionID: s.SessionPart.SID}, nil
}

func (a *Agent) RunA(ctx context.Context, message Message, handlerFunc StreamFunc, opts ...adk.AgentRunOption) error {
	resp := a.Run(ctx, []*schema.Message{{Role: message.Role, Content: message.Content}}, opts...)
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

		if sm := e.Output.MessageOutput.MessageStream; sm != nil {
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
					handlerFunc(Message{Type: ContentType, Content: m.Content})
				}
				if m.ReasoningContent != "" {
					handlerFunc(Message{Type: ThinkingType, Content: m.ReasoningContent})
				}
			}
		}

		if tm := e.Output.MessageOutput.ToolName; tm != "" {
			handlerFunc(Message{Type: ToolType, Content: tm})
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
