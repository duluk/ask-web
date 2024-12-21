package summarize

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func Summarize(contents []string, query string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("environment variable OPENAI_API_KEY must be set")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	prompt := buildPrompt(contents, query)

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4oMini,
		MaxTokens: 300,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
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

func buildPrompt(contents []string, query string) string {
	prompt := fmt.Sprintf("Please provide a detailed summary of the following text that is related to the query '%s'. ", query)

	for _, content := range contents {
		prompt += "\n" + content
	}

	prompt += "\nSummary:"

	return prompt
}
