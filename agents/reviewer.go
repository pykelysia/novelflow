package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
)

// This is the sub agent that focuses on reviewing the complated chapter and fix it whitout change any real content.

func createReviewrAgent(ctx context.Context) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{})
}
