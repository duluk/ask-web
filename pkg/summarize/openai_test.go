package summarize

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sashabaranov/go-openai"

	"ask-web/pkg/config"
)

type mockOpenAIModel struct {
	createChatCompletionFunc func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func (m *mockOpenAIModel) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.createChatCompletionFunc(ctx, request)
}

func newTestOpenAISummarizer(opts *config.Opts, mockClient OpenAIModel) *OpenAISummarizer {
	return &OpenAISummarizer{
		client:       mockClient,
		opts:         opts,
		systemPrompt: fmt.Sprintf("Fit the response within %d tokens", opts.MaxTokens),
	}
}

func TestOpenAISummarizer(t *testing.T) {
	testCases := []struct {
		name            string
		contents        []string
		query           string
		maxTokens       int
		mockResponse    openai.ChatCompletionResponse
		mockError       error
		expectedSummary string
		expectedError   error
	}{
		{
			name:      "Successful Summary",
			contents:  []string{"This is the first content.", "This is the second content."},
			query:     "Test query",
			maxTokens: 100,
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "This is a test summary.",
						},
					},
				},
			},
			expectedSummary: "This is a test summary.",
			expectedError:   nil,
		},
		{
			name:      "No Summary Generated",
			contents:  []string{"Some content."},
			query:     "Another query",
			maxTokens: 50,
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{},
			},
			expectedError: errors.New("no summary generated"),
		},
		{
			name:          "API Error",
			contents:      []string{"Content here."},
			query:         "Error query",
			maxTokens:     150,
			mockError:     errors.New("API request failed"),
			expectedError: errors.New("API request failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockOpenAIModel{
				createChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
					return tc.mockResponse, tc.mockError
				},
			}

			opts := &config.Opts{
				SummaryPrompt: "Please provide a detailed summary of the following text that is directly related to the query",
				MaxTokens:     tc.maxTokens,
			}

			summarizer := newTestOpenAISummarizer(opts, mockClient)
			summary, err := summarizer.Summarize(context.Background(), tc.contents, tc.query)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error '%v', got '%v'", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if summary != tc.expectedSummary {
				t.Errorf("Expected summary '%s', got '%s'", tc.expectedSummary, summary)
			}
		})
	}
}
