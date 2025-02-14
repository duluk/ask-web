package summarize

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"ask-web/pkg/config"
)

type OpenAIModel interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

type OpenAISummarizer struct {
	client       OpenAIModel
	opts         *config.Opts
	systemPrompt string
}

func NewOpenAISummarizer(apiKey string, opts *config.Opts) (*OpenAISummarizer, error) {
	client := openai.NewClient(apiKey)
	return &OpenAISummarizer{
		client:       client,
		opts:         opts,
		systemPrompt: fmt.Sprintf("Fit the response within %d tokens", opts.MaxTokens),
	}, nil
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, contents []string, query string) (string, error) {
	prompt := buildPrompt(contents, query, s.opts.SummaryPrompt)

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		MaxTokens:   s.opts.MaxTokens,
		Temperature: float32(s.opts.Temperature),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no summary generated")
	}

	return resp.Choices[0].Message.Content, nil
}
