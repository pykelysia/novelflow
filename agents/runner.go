package agents

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
)

type AgentRunner struct {
	*adk.Runner
}

type AgentRunnerConfig struct {
	*deep.Config
}

type Message struct {
	Type    string
	Content string
}

type StreamFunc func(Message) bool

func NewAgentRunner(ctx context.Context, config *AgentRunnerConfig) (*AgentRunner, error) {
	var err error = nil

	if config.ModelRetryConfig == nil {
		config.ModelRetryConfig = &adk.ModelRetryConfig{
			MaxRetries: 3,
			IsRetryAble: func(ctx context.Context, err error) bool {
				return strings.Contains(err.Error(), "429") ||
					strings.Contains(err.Error(), "Too Many Request")
			},
		}
	}

	backend, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		return nil, err
	}

	if config.Backend == nil {
		config.Backend = backend
	}
	if config.StreamingShell == nil {
		config.StreamingShell = backend
	}

	a, err := deep.New(ctx, config.Config)
	if err != nil {
		return nil, err
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a,
		EnableStreaming: true,
	})

	return &AgentRunner{
		Runner: r,
	}, nil
}

func (ar *AgentRunner) RunA(ctx context.Context, messages []adk.Message, handlerFunc StreamFunc, opts ...adk.AgentRunOption) error {
	resp := ar.Run(ctx, messages, opts...)
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
		if sm != nil {
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
						Type:    "Content",
						Content: m.Content,
					})
				}
				if m.ReasoningContent != "" {
					handlerFunc(Message{
						Type:    "Thinking",
						Content: m.ReasoningContent,
					})
				}
			}
		}
		tm := e.Output.MessageOutput.ToolName
		if tm != "" {
			handlerFunc(Message{
				Type:    "tool",
				Content: tm,
			})
		}
	}
	return nil
}
