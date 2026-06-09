package core

import (
	"context"
	"fmt"
	"strings"

	"novelflow/agents/session"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

const compressedKey = "_nf_compressed"

// CompressConfig holds tunable parameters for context compression.
// Zero values fall back to the defaults below.
type CompressConfig struct {
	ThresholdChars  int // trigger compression when total chars exceed this
	KeepRecentMsgs  int // number of most-recent messages to preserve verbatim
	KeepRecentTools int // number of most-recent tool results to preserve verbatim
}

func (c *CompressConfig) withDefaults() CompressConfig {
	out := *c
	if out.ThresholdChars <= 0 {
		out.ThresholdChars = viper.GetInt("context_compress.threshold_chars")
		if out.ThresholdChars <= 0 {
			out.ThresholdChars = 600000
		}
	}
	if out.KeepRecentMsgs <= 0 {
		out.KeepRecentMsgs = viper.GetInt("context_compress.keep_recent_msgs")
		if out.KeepRecentMsgs <= 0 {
			out.KeepRecentMsgs = 10
		}
	}
	if out.KeepRecentTools <= 0 {
		out.KeepRecentTools = viper.GetInt("context_compress.keep_recent_tools")
		if out.KeepRecentTools <= 0 {
			out.KeepRecentTools = 5
		}
	}
	return out
}

// CompressMiddleware replaces Eino's built-in summarization middleware with a
// custom implementation that:
//   - preserves recent messages verbatim
//   - archives old tool results to MongoDB
//   - generates a structured summary (user intent / progress / next steps / rules)
type CompressMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	sess      *session.Session
	liteModel model.BaseChatModel
	cfg       CompressConfig
}

func newCompressMiddleware(ctx context.Context, sess *session.Session) (*CompressMiddleware, error) {
	lm, err := GetLiteChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("compress middleware: get lite model: %w", err)
	}
	return &CompressMiddleware{
		sess:      sess,
		liteModel: lm,
		cfg:       CompressConfig{},
	}, nil
}

// BeforeModelRewriteState runs after SessionMiddleware has injected history
// (middleware BeforeModel hooks execute in reverse-append order in Eino ADK).
// It checks the total character count and, if over threshold, compresses.
func (m *CompressMiddleware) BeforeModelRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	_ *adk.ModelContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	if _, found, _ := adk.GetRunLocalValue(ctx, compressedKey); found {
		return ctx, state, nil
	}
	_ = adk.SetRunLocalValue(ctx, compressedKey, true)

	cfg := m.cfg.withDefaults()

	if totalChars(state.Messages) <= cfg.ThresholdChars {
		return ctx, state, nil
	}

	// Separate system messages from conversation messages.
	var sysMsgs, convMsgs []*schema.Message
	for _, msg := range state.Messages {
		if msg.Role == schema.System {
			sysMsgs = append(sysMsgs, msg)
		} else {
			convMsgs = append(convMsgs, msg)
		}
	}

	if len(convMsgs) <= cfg.KeepRecentMsgs {
		return ctx, state, nil
	}

	oldMsgs := convMsgs[:len(convMsgs)-cfg.KeepRecentMsgs]
	recentMsgs := convMsgs[len(convMsgs)-cfg.KeepRecentMsgs:]

	// Identify tool-result messages in old and apply KeepRecentTools.
	archives, oldMsgsFiltered := archiveOldToolResults(
		ctx, m.sess, oldMsgs, cfg.KeepRecentTools,
	)

	// Generate structured summary via lite LLM.
	summaryText, err := m.generateSummary(ctx, sysMsgs, oldMsgsFiltered, archives)
	if err != nil {
		// On failure, skip compression rather than breaking the agent.
		return ctx, state, nil
	}

	summaryMsg := session.Message{
		Type:    session.SummaryType,
		Role:    schema.Assistant,
		Content: summaryText,
	}

	// Persist summary and archives; update session rawCache.
	recentStorage := make([]session.Message, 0, len(recentMsgs))
	for _, rm := range recentMsgs {
		recentStorage = append(recentStorage, schemaToMessage(rm))
	}
	if err := m.sess.Compress(ctx, summaryMsg, archives, recentStorage); err != nil {
		return ctx, state, nil
	}

	// Rebuild state messages: [system...] + [summary] + [recent...]
	newMsgs := make([]*schema.Message, 0, len(sysMsgs)+1+len(recentMsgs))
	newMsgs = append(newMsgs, sysMsgs...)
	newMsgs = append(newMsgs, &schema.Message{Role: schema.Assistant, Content: summaryText})
	newMsgs = append(newMsgs, recentMsgs...)
	state.Messages = newMsgs

	// Update the savedLen cursor so SessionMiddleware.AfterModel saves only
	// genuinely new messages after the model responds.
	_ = adk.SetRunLocalValue(ctx, savedLenKey, len(state.Messages))

	return ctx, state, nil
}

// archiveOldToolResults scans oldMsgs for tool results, keeps the last
// keepRecentTools verbatim, and archives the rest to MongoDB.
// It returns the archives and the filtered oldMsgs (tool results replaced by
// concise stubs so the summary LLM still knows what tools were called).
func archiveOldToolResults(
	ctx context.Context,
	sess *session.Session,
	oldMsgs []*schema.Message,
	keepRecentTools int,
) ([]session.ToolResultArchive, []*schema.Message) {
	// Collect tool-result indices in old messages (role == Tool).
	var toolResultIdx []int
	for i, m := range oldMsgs {
		if m.Role == schema.Tool {
			toolResultIdx = append(toolResultIdx, i)
		}
	}

	archiveUntil := len(toolResultIdx) - keepRecentTools
	if archiveUntil <= 0 {
		return nil, oldMsgs
	}

	toArchive := make(map[int]bool, archiveUntil)
	for _, idx := range toolResultIdx[:archiveUntil] {
		toArchive[idx] = true
	}

	var archives []session.ToolResultArchive
	filtered := make([]*schema.Message, 0, len(oldMsgs))

	// We need the preceding assistant message (tool call) to get args.
	for i, m := range oldMsgs {
		if !toArchive[i] {
			filtered = append(filtered, m)
			continue
		}
		// Find matching tool call args from previous assistant message.
		toolArgs := ""
		if i > 0 && len(oldMsgs[i-1].ToolCalls) > 0 {
			toolArgs = oldMsgs[i-1].ToolCalls[0].Function.Arguments
		}
		arch := session.NewArchive(
			sess.SessionPart.SID,
			m.ToolName,
			toolArgs,
			m.Content,
			true, // treat all persisted results as successful
		)
		archives = append(archives, arch)

		// Replace with a terse stub so the summary LLM retains tool call context.
		stub := fmt.Sprintf("[归档工具结果 archive_id=%s] %s → 已归档", arch.ID, m.ToolName)
		filtered = append(filtered, &schema.Message{
			Role:       schema.Tool,
			Content:    stub,
			ToolCallID: m.ToolCallID,
			ToolName:   m.ToolName,
		})
	}

	return archives, filtered
}

const summaryPromptHeader = `你是对话历史压缩助手。请根据系统提示词和对话历史，严格按照以下格式输出摘要，不要输出其他任何内容：

## 用户意图
[描述用户最终想要完成的总体目标]

## 当前进度
[描述目前已经完成的工作和成果，尽量具体]

## 下一步计划
[描述接下来需要执行的具体步骤]

## 用户规则
[从系统提示词中提取用户明确要求的规则、约束和偏好，如无则写"无特殊规则"]

`

func (m *CompressMiddleware) generateSummary(
	ctx context.Context,
	sysMsgs []*schema.Message,
	msgs []*schema.Message,
	archives []session.ToolResultArchive,
) (string, error) {
	var sb strings.Builder
	sb.WriteString(summaryPromptHeader)
	if len(sysMsgs) > 0 {
		sb.WriteString("--- 系统提示词（含用户规则）---\n")
		for _, m := range sysMsgs {
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("--- 以下为需要压缩的对话历史 ---\n")
	sb.WriteString(renderMessages(msgs))
	prompt := sb.String()

	resp, err := m.liteModel.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: prompt},
	})
	if err != nil {
		return "", err
	}

	summaryText := resp.Content

	if len(archives) > 0 {
		var sb strings.Builder
		sb.WriteString(summaryText)
		sb.WriteString("\n\n## 已归档工具调用\n")
		for _, a := range archives {
			status := "成功"
			if !a.Success {
				status = "失败"
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s → %s\n", a.ID, a.ToolName, status))
		}
		summaryText = sb.String()
	}

	return summaryText, nil
}

func totalChars(msgs []*schema.Message) int {
	n := 0
	for _, m := range msgs {
		n += len(m.Content)
		n += len(m.ReasoningContent)
		for _, tc := range m.ToolCalls {
			n += len(tc.Function.Name) + len(tc.Function.Arguments)
		}
	}
	return n
}

func renderMessages(msgs []*schema.Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		role := string(m.Role)
		if len(m.ToolCalls) > 0 {
			tc := m.ToolCalls[0]
			sb.WriteString(fmt.Sprintf("[%s] 调用工具 %s，参数：%s\n", role, tc.Function.Name, tc.Function.Arguments))
		} else if m.Role == schema.Tool {
			sb.WriteString(fmt.Sprintf("[tool:%s] %s\n", m.ToolName, m.Content))
		} else {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", role, m.Content))
		}
	}
	return sb.String()
}
