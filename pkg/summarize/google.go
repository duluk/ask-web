package summarize

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"ask-web/pkg/config"
)

type GenerativeModel interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
	SetTemperature(float32)
	SetMaxOutputTokens(int32)
}

type GoogleSummarizer struct {
	client       *genai.Client
	model        GenerativeModel // Changed from *genai.GenerativeModel
	opts         *config.Opts
	systemPrompt string
}

func NewGoogleSummarizer(apiKey string, opts *config.Opts) (*GoogleSummarizer, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google AI client: %w", err)
	}

	rawModel := client.GenerativeModel("gemini-2.0-flash-001")
	var model GenerativeModel = rawModel

	model.SetTemperature(float32(opts.Temperature))
	model.SetMaxOutputTokens(int32(opts.MaxTokens)) // Cast to int32

	return &GoogleSummarizer{
		client:       client,
		model:        model,
		opts:         opts,
		systemPrompt: fmt.Sprintf("Fit the response within %d tokens", opts.MaxTokens),
	}, nil
}

func (s *GoogleSummarizer) Summarize(ctx context.Context, contents []string, query string) (string, error) {
	prompt := buildPrompt(contents, query, s.opts.SummaryPrompt)

	resp, err := s.model.GenerateContent(ctx, genai.Text(s.systemPrompt+"\n\n"+prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no summary generated")
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
