package core

import (
	"context"
	"encoding/gob"

	"novelflow/agents/session"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

const savedLenKey = "_nf_session_saved_len"

func init() {
	// SetRunLocalValue stores values in map[string]any via gob; int must be registered.
	gob.Register(int(0))
}

// SessionMiddleware persists conversation messages to a MessageStore via the
// BeforeModelRewriteState / AfterModelRewriteState / AfterToolCallsRewriteState hooks.
// It injects history into the state on the first model call, saves new messages
// after each model response and after each round of tool calls.
type SessionMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	store session.MessageStore
}

func newSessionMiddleware(store session.MessageStore) *SessionMiddleware {
	return &SessionMiddleware{store: store}
}

// BeforeModelRewriteState runs only on the first model call per agent run.
// It injects stored history between the system message and the current user
// message, then persists the user message.
func (m *SessionMiddleware) BeforeModelRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	_ *adk.ModelContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	if _, found, _ := adk.GetRunLocalValue(ctx, savedLenKey); found {
		return ctx, state, nil
	}

	history := m.store.Load()
	if len(history) > 0 {
		// Split at the boundary between system messages and the rest.
		split := 0
		for split < len(state.Messages) && state.Messages[split].Role == schema.System {
			split++
		}
		msgs := make([]*schema.Message, 0, len(state.Messages)+len(history))
		msgs = append(msgs, state.Messages[:split]...)
		msgs = append(msgs, history...)
		msgs = append(msgs, state.Messages[split:]...)
		state.Messages = msgs
	}

	// Save the incoming user message (last message after injection).
	if n := len(state.Messages); n > 0 {
		if last := state.Messages[n-1]; last.Role == schema.User {
			_ = m.store.Append(schemaToMessage(last))
		}
	}

	_ = adk.SetRunLocalValue(ctx, savedLenKey, len(state.Messages))
	return ctx, state, nil
}

// AfterModelRewriteState saves any messages added by the model response.
func (m *SessionMiddleware) AfterModelRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	_ *adk.ModelContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	m.persistNewMessages(ctx, state.Messages)
	return ctx, state, nil
}

// AfterToolCallsRewriteState saves tool call results appended after the model response.
func (m *SessionMiddleware) AfterToolCallsRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	_ *adk.ToolCallsContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	m.persistNewMessages(ctx, state.Messages)
	return ctx, state, nil
}

func (m *SessionMiddleware) persistNewMessages(ctx context.Context, messages []*schema.Message) {
	raw, found, _ := adk.GetRunLocalValue(ctx, savedLenKey)
	if !found {
		return
	}
	savedLen, ok := raw.(int)
	if !ok || savedLen >= len(messages) {
		return
	}
	for _, msg := range messages[savedLen:] {
		_ = m.store.Append(schemaToMessage(msg))
	}
	_ = adk.SetRunLocalValue(ctx, savedLenKey, len(messages))
}

// schemaToMessage converts a schema.Message to the storage format used by MessageStore.
// The encoding mirrors the format expected by Session.Load() for round-trip fidelity.
func schemaToMessage(m *schema.Message) session.Message {
	if m.Role == schema.Tool {
		return session.Message{Type: session.ToolResultType, Role: schema.Tool, Content: m.ToolName, ToolResult: m.Content}
	}
	if len(m.ToolCalls) > 0 {
		tc := m.ToolCalls[0]
		return session.Message{Type: session.ToolResultType, Role: schema.Assistant, Content: tc.Function.Name + "\n" + tc.Function.Arguments}
	}
	return session.Message{Type: session.ContentType, Role: m.Role, Content: m.Content}
}
