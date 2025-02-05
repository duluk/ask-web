package summarize

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"ask-web/pkg/config"
)

type OpenAIClient interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// client is for testing purposes
func Summarize(opts *config.Opts, apiKey string, contents []string, query string, client OpenAIClient) (string, error) {
	if client == nil {
		client = openai.NewClient(apiKey)
	}
	ctx := context.Background()

	systemPrompt := fmt.Sprintf("Fit the response within %d tokens", opts.MaxTokens)
	prompt := buildPrompt(contents, query, opts.SummaryPrompt)

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4oMini,
		MaxTokens: opts.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no summary generated")
	}

	return resp.Choices[0].Message.Content, nil
}

func buildPrompt(contents []string, query string, summaryPrompt string) string {
	prompt := fmt.Sprintf("%s '%s'. ", summaryPrompt, query)

	for _, content := range contents {
		prompt += "\n" + content
	}

	prompt += "\nSummary:"

	return prompt
}
