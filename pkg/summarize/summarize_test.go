package summarize

import (
	"context"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"

	"ask-web/pkg/config"
)

// mockOpenAIClient is a mock implementation of the OpenAI client.
type mockOpenAIClient struct {
	createChatCompletionFunc func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.createChatCompletionFunc(ctx, request)
}

func TestOpenAISummarizer(t *testing.T) {
	testCases := []struct {
		name            string
		apiKey          string
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
			apiKey:    "test-api-key",
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
			apiKey:    "test-api-key",
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
			apiKey:        "test-api-key",
			contents:      []string{"Content here."},
			query:         "Error query",
			maxTokens:     150,
			mockError:     errors.New("API request failed"),
			expectedError: errors.New("API request failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockOpenAIClient{
				createChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
					return tc.mockResponse, tc.mockError
				},
			}

			opts := &config.Opts{
				SummaryPrompt: "Please provide a detailed summary of the following text that is directly related to the query",
				MaxTokens:     tc.maxTokens,
			}

			summarizer := NewOpenAISummarizer(tc.apiKey, opts, mockClient)
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

func TestBuildPrompt(t *testing.T) {
	testCases := []struct {
		name     string
		contents []string
		query    string
		expected string
	}{
		{
			name:     "Single Content",
			contents: []string{"This is a test content."},
			query:    "Test query",
			expected: "Please provide a detailed summary of the following text that is related to the query 'Test query'. \nThis is a test content.\nSummary:",
		},
		{
			name:     "Multiple Contents",
			contents: []string{"First content.", "Second content."},
			query:    "Another query",
			expected: "Please provide a detailed summary of the following text that is related to the query 'Another query'. \nFirst content.\nSecond content.\nSummary:",
		},
		{
			name:     "Empty Contents",
			contents: []string{},
			query:    "Empty query",
			expected: "Please provide a detailed summary of the following text that is related to the query 'Empty query'. \nSummary:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sp := "Please provide a detailed summary of the following text that is related to the query"
			prompt := buildPrompt(tc.contents, tc.query, sp)
			if prompt != tc.expected {
				t.Errorf("Expected prompt '%s', got '%s'", tc.expected, prompt)
			}
		})
	}
}
