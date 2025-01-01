package search

import (
	"testing"
)

// Test search for DDG since no API key is required
func TestDDGSearch(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		maxResults int
		wantErr    bool
	}{
		{
			name:       "Basic search",
			query:      "golang programming",
			maxResults: 3,
			wantErr:    false,
		},
		{
			name:       "Custom options",
			query:      "golang programming",
			maxResults: 3,
			wantErr:    false,
		},
		{
			name:       "Empty query",
			query:      "",
			maxResults: 3,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := DDGSearch(tt.query, tt.maxResults)

			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(results) == 0 {
					t.Error("Search() returned no results")
				}

				if len(results) > tt.maxResults {
					t.Errorf("Search() returned more results than MaxResults: got %d, want <= %d", len(results), tt.maxResults)
				}

				// Check that results are properly formatted
				for _, result := range results {
					if result.Title == "" {
						t.Error("Search() returned result with empty title")
					}
					if result.URL == "" {
						t.Error("Search() returned result with empty link")
					}
				}
			}
		})
	}
}
