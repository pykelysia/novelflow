package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"novelflow/database/mongodb"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// GenerationIntent 用于从 service 层传递生成意图，供 rule 筛选使用。
type GenerationIntent struct {
	Genre        string
	Concept      string
	Style        string
	Requirements string
}

type ruleSelectionResult struct {
	SelectedIDs []string `json:"selected_ids"`
}

// RunWithRules 在调用 NewMainAgent 前执行意图分析，筛选并注入相关 rules。
func RunWithRules(ctx context.Context, sessionID string, userID uint, enabledRules []mongodb.Rule, intent *GenerationIntent) (*Agent, error) {
	rulesContent := buildRulesContent(ctx, enabledRules, intent)
	return NewMainAgent(ctx, sessionID, userID, rulesContent)
}

// buildRulesContent 通过 Eino Chain 执行规则筛选，失败时降级注入全部规则。
func buildRulesContent(ctx context.Context, rules []mongodb.Rule, intent *GenerationIntent) string {
	if len(rules) == 0 {
		return ""
	}
	content, err := runRuleSelectionChain(ctx, rules, intent)
	if err != nil {
		log.Printf("[rules] 规则筛选失败，降级注入全部规则: %v", err)
		return formatRulesContent(rules)
	}
	return content
}

// runRuleSelectionChain 构建并运行规则筛选 Chain，返回 rulesContent 字符串。
// Chain 数据流：map[string]any → []*schema.Message → *schema.Message → string
// 任何节点失败都向上返回 error，由 buildRulesContent 统一降级。
func runRuleSelectionChain(ctx context.Context, rules []mongodb.Rule, intent *GenerationIntent) (string, error) {
	idIndex := buildIDIndex(rules)

	cm, err := getLiteChatModel(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 lite_llm 失败: %w", err)
	}

	tpl := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(ruleSelectionSystemPrompt),
		schema.UserMessage(ruleSelectionUserPromptTpl),
	)

	parseLambda := compose.InvokableLambda(func(_ context.Context, msg *schema.Message) (string, error) {
		return parseSelectionResponse(msg.Content, rules, idIndex)
	})

	chain := compose.NewChain[map[string]any, string]()
	chain.AppendChatTemplate(tpl).AppendChatModel(cm).AppendLambda(parseLambda)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return "", fmt.Errorf("chain 编译失败: %w", err)
	}

	return runnable.Invoke(ctx, buildSelectionVars(rules, intent))
}

// buildIDIndex 构建 id → Rule 查找表。
func buildIDIndex(rules []mongodb.Rule) map[string]mongodb.Rule {
	m := make(map[string]mongodb.Rule, len(rules))
	for _, r := range rules {
		m[r.ID.Hex()] = r
	}
	return m
}

// buildSelectionVars 构建 ChatTemplate 所需的模板变量 map。
func buildSelectionVars(rules []mongodb.Rule, intent *GenerationIntent) map[string]any {
	var sb strings.Builder
	for _, r := range rules {
		fmt.Fprintf(&sb, "- ID: %s | 名称: %s | 内容摘要: %s\n",
			r.ID.Hex(), r.Name, truncate(r.Content, 200))
	}
	if intent == nil {
		intent = &GenerationIntent{}
	}
	return map[string]any{
		"genre":        intent.Genre,
		"concept":      intent.Concept,
		"style":        intent.Style,
		"requirements": intent.Requirements,
		"rule_lines":   sb.String(),
	}
}

// parseSelectionResponse 解析 LLM 返回的 JSON，失败时降级返回全量规则的格式化文本。
func parseSelectionResponse(raw string, allRules []mongodb.Rule, idIndex map[string]mongodb.Rule) (string, error) {
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		log.Printf("[rules] LLM 返回格式无效，降级注入全部规则")
		return formatRulesContent(allRules), nil
	}

	var result ruleSelectionResult
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		log.Printf("[rules] JSON 解析失败，降级注入全部规则: %v", err)
		return formatRulesContent(allRules), nil
	}

	selected := make([]mongodb.Rule, 0, len(result.SelectedIDs))
	for _, id := range result.SelectedIDs {
		if r, ok := idIndex[id]; ok {
			selected = append(selected, r)
		}
	}
	if len(selected) == 0 {
		return formatRulesContent(allRules), nil
	}
	return formatRulesContent(selected), nil
}

// formatRulesContent 将规则列表格式化为系统提示注入字符串。
func formatRulesContent(rules []mongodb.Rule) string {
	if len(rules) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("以下规则本次创作必须严格遵守（优先级高于一般写作建议）：\n\n")
	for i, r := range rules {
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, r.Name, r.Content)
	}
	return sb.String()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
