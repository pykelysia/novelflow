package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func singleChunkReader(msg string) *schema.StreamReader[string] {
	r, w := schema.Pipe[string](1)
	_ = w.Send(msg, nil)
	w.Close()
	return r
}

func safeWarpReader(sr *schema.StreamReader[string]) *schema.StreamReader[string] {
	r, w := schema.Pipe[string](1)
	go func() {
		defer w.Close()
		for {
			chunk, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				_ = w.Send(fmt.Sprintf("\n[tool error] %v", err), nil)
				return
			}
			_ = w.Send(chunk, nil)
		}
	}()
	return r
}

func isRetryAble(ctx context.Context, err error) bool {
	return strings.Contains(err.Error(), "429") ||
		strings.Contains(err.Error(), "Too Many Request")
}

func defaultUnknownToolHandler(ctx context.Context, name, input string) (string, error) {
	return fmt.Sprintf("[tool error]: tool %s is not available. Please choose another tool.", name), nil
}
