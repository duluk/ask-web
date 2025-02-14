package summarize

import (
	"context"
	"fmt"
)

type Summarizer interface {
	Summarize(ctx context.Context, contents []string, query string) (string, error)
}

func buildPrompt(contents []string, query string, summaryPrompt string) string {
	prompt := fmt.Sprintf("%s '%s'. ", summaryPrompt, query)

	for _, content := range contents {
		prompt += "\n" + content
	}

	prompt += "\nSummary:"

	return prompt
}
