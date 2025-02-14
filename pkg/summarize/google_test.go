package summarize

import (
	"context"
	"errors"
	"testing"

	"github.com/google/generative-ai-go/genai"

	"ask-web/pkg/config"
)

type mockGoogleResponse struct {
	candidates []*genai.Candidate
	err        error
}

type mockGoogleModel struct {
	response mockGoogleResponse
}

func (m *mockGoogleModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	if m.response.err != nil {
		return nil, m.response.err
	}
	return &genai.GenerateContentResponse{
		Candidates: m.response.candidates,
	}, nil
}

func (m *mockGoogleModel) SetTemperature(float32)   {}
func (m *mockGoogleModel) SetMaxOutputTokens(int32) {} // Changed from int to int32

// Remove all the other Set* methods as they're not part of our interface
// func (m *mockGoogleModel) SetTopP(float32)            {}
// func (m *mockGoogleModel) SetTopK(int32)              {}
// func (m *mockGoogleModel) SetCandidateCount(int32)    {}
// func (m *mockGoogleModel) SetStopSequences([]string)  {}
// func (m *mockGoogleModel) SetSafetySettings([]*genai.SafetySetting) {}

func TestGoogleSummarizer_Summarize(t *testing.T) {
	testCases := []struct {
		name            string
		contents        []string
		query           string
		maxTokens       int
		mockResponse    mockGoogleResponse
		expectedSummary string
		expectedError   error
	}{
		{
			name:      "Successful Summary",
			contents:  []string{"This is the first content.", "This is the second content."},
			query:     "Test query",
			maxTokens: 100,
			mockResponse: mockGoogleResponse{
				candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("This is a test summary."),
							},
						},
					},
				},
			},
			expectedSummary: "This is a test summary.",
			expectedError:   nil,
		},
		{
			name:     "No Summary Generated",
			contents: []string{"Some content."},
			query:    "Another query",
			mockResponse: mockGoogleResponse{
				candidates: []*genai.Candidate{},
			},
			expectedError: errors.New("no summary generated"),
		},
		{
			name:     "API Error",
			contents: []string{"Content here."},
			query:    "Error query",
			mockResponse: mockGoogleResponse{
				err: errors.New("API request failed"),
			},
			expectedError: errors.New("API request failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockModel := &mockGoogleModel{
				response: tc.mockResponse,
			}

			summarizer := &GoogleSummarizer{
				model: mockModel,
				opts: &config.Opts{
					MaxTokens:     tc.maxTokens,
					SummaryPrompt: "Please provide a detailed summary of the following text that is directly related to the query",
				},
			}

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
