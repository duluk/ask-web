package summarize

import (
	"testing"
)

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
