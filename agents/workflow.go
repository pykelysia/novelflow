package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"novelflow/database/mongodb"

	"github.com/cloudwego/eino/schema"
)

// GenerationIntent 用于从 service 层传递生成意图，供 rule 筛选使用。
type GenerationIntent struct {
	Genre        string
	Concept      string
	Style        string
	Requirements string
}

// RunWithRules 在调用 NewMainAgent 前执行意图分析，筛选并注入相关 rules。
func RunWithRules(ctx context.Context, sessionID string, userID uint, enabledRules []mongodb.Rule, intent *GenerationIntent) (*Agent, error) {
	rulesContent := buildRulesContent(ctx, enabledRules, intent)
	return NewMainAgent(ctx, sessionID, userID, rulesContent)
}

// buildRulesContent 用 lite_llm 筛选相关规则，失败时降级注入全部规则。
func buildRulesContent(ctx context.Context, rules []mongodb.Rule, intent *GenerationIntent) string {
	if len(rules) == 0 {
		return ""
	}

	selected := selectRules(ctx, rules, intent)
	if len(selected) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("以下规则本次创作必须严格遵守（优先级高于一般写作建议）：\n\n")
	for i, r := range selected {
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, r.Name, r.Content)
	}
	return sb.String()
}

type ruleSelectionResult struct {
	SelectedIDs []string `json:"selected_ids"`
}

// selectRules 调用 lite_llm 根据请求意图筛选规则，失败时返回全部规则。
func selectRules(ctx context.Context, rules []mongodb.Rule, intent *GenerationIntent) []mongodb.Rule {
	cm, err := getLiteChatModel(ctx)
	if err != nil {
		log.Printf("[rules] 获取 lite_llm 失败，降级注入全部规则: %v", err)
		return rules
	}

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: `你是写作规则筛选助手。根据小说创作请求，从规则列表中找出最相关的规则 ID。只返回 JSON：{"selected_ids": ["id1", "id2"]}，不输出任何其他内容。`,
	}

	var ruleLines strings.Builder
	idIndex := make(map[string]mongodb.Rule, len(rules))
	for _, r := range rules {
		id := r.ID.Hex()
		fmt.Fprintf(&ruleLines, "- ID: %s | 名称: %s | 内容摘要: %s\n", id, r.Name, truncate(r.Content, 200))
		idIndex[id] = r
	}

	userContent := fmt.Sprintf(
		"## 本次创作请求\n\n类型：%s\n概念：%s\n风格：%s\n其他要求：%s\n\n## 用户规则列表\n\n%s\n请返回相关规则 ID 的 JSON。",
		intent.Genre, intent.Concept, intent.Style, intent.Requirements, ruleLines.String(),
	)
	userMsg := &schema.Message{Role: schema.User, Content: userContent}

	resp, err := cm.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		log.Printf("[rules] 意图分析 LLM 调用失败，降级注入全部规则: %v", err)
		return rules
	}

	raw := strings.TrimSpace(resp.Content)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		log.Printf("[rules] 意图分析返回格式无效，降级注入全部规则")
		return rules
	}

	var result ruleSelectionResult
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		log.Printf("[rules] 意图分析 JSON 解析失败，降级注入全部规则: %v", err)
		return rules
	}

	selected := make([]mongodb.Rule, 0, len(result.SelectedIDs))
	for _, id := range result.SelectedIDs {
		if r, ok := idIndex[id]; ok {
			selected = append(selected, r)
		}
	}
	if len(selected) == 0 {
		return rules
	}
	return selected
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
