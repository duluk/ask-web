package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"ask-web/pkg/config"
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

type APIKeys struct {
	GoogleAPIKey  string
	GoogleCSEID   string
	BingAPIKey    string
	BingConfigKey string
	OpenAIKey     string
}

type FilterFunc func(SearchResult) bool

const MaxTimeoutSeconds = 10
const ExtraResultsFactor = 2.0

// Take a query argument and then send it to the OpenAI API to generate a concise search query to use
func CreateSearchQuery(opts *config.Opts, apiKey string, query string) (string, error) {
	client := openai.NewClient(apiKey)

	systemPrompt := "You are generating a query to pass to a search engine. Return only the query, do not generate extraneous information. Try not to include dates unless in the query itself; your knowledge base it cutoff and you may get it wrong."
	prompt := fmt.Sprintf("%s: '%s'", opts.QueryPrompt, query)

	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		MaxTokens:   50,
		Temperature: 0.3,
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
		return "", errors.New("no query generated")
	}

	return resp.Choices[0].Message.Content, nil
}
