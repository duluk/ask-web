package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	// "github.com/adrg/xdg"

	"ask-web/pkg/search"
)

func TestCleanText(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No HTML, no extra space",
			input:    "This is a test string.",
			expected: "This is a test string.",
		},
		{
			name:     "HTML tags",
			input:    "<p>This is <b>a test</b> string.</p>",
			expected: "This is a test string.",
		},
		{
			name:     "Extra leading and trailing spaces",
			input:    "   This is a test string.   ",
			expected: "This is a test string.",
		},
		{
			name:     "Multiple internal spaces",
			input:    "This  is   a    test string.",
			expected: "This is a test string.",
		},
		{
			name:     "Mixed HTML and spaces",
			input:    "  <p> This  is   <b>a</b>  test </p>  string.  ",
			expected: "This is a test string.",
		},
		{
			name:     "Handle newlines",
			input:    "This is a\n test \nstring.",
			expected: "This is a test string.",
		},
		{
			name:     "Handle non-breaking spaces",
			input:    "This is a &nbsp;test string.",
			expected: "This is a test string.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := CleanText(tc.input)
			if actual != tc.expected {
				t.Errorf("CleanText(%q) = %q; want %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestDedupeResults(t *testing.T) {
	testCases := []struct {
		name     string
		input    []search.SearchResult
		expected []search.SearchResult
	}{
		{
			name: "No duplicates",
			input: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2"},
				{URL: "https://example.com/page3"},
			},
			expected: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2"},
				{URL: "https://example.com/page3"},
			},
		},
		{
			name: "Simple duplicates",
			input: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2"},
			},
			expected: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2"},
			},
		},
		{
			name: "Duplicates with query params",
			input: []search.SearchResult{
				{URL: "https://example.com/page1?q=test"},
				{URL: "https://example.com/page1?q=another"},
				{URL: "https://example.com/page2"},
			},
			expected: []search.SearchResult{
				{URL: "https://example.com/page1?q=test"},
				{URL: "https://example.com/page2"},
			},
		},
		{
			name: "Mixed duplicates and no duplicates",
			input: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2?q=test"},
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page3"},
				{URL: "https://example.com/page2?q=another"},
			},
			expected: []search.SearchResult{
				{URL: "https://example.com/page1"},
				{URL: "https://example.com/page2?q=test"},
				{URL: "https://example.com/page3"},
			},
		},
		{
			name:     "Empty input",
			input:    []search.SearchResult{},
			expected: []search.SearchResult{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := DedupeResults(tc.input)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("DedupeResults(%v) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestSetupKeys(t *testing.T) {
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	tempDir, err := os.MkdirTemp("", "ask-web-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("XDG_CONFIG_HOME", tempDir)

	askWebConfigDir := filepath.Join(tempDir, "ask-web")
	err = os.MkdirAll(askWebConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create ask-web config dir: %v", err)
	}

	testCases := []struct {
		name     string
		envVars  map[string]string
		files    map[string]string
		expected search.APIKeys
	}{
		{
			name: "All keys from environment",
			envVars: map[string]string{
				"GOOGLE_API_KEY":  "env_google_api_key",
				"GOOGLE_CSE_ID":   "env_google_cse_id",
				"BING_API_KEY":    "env_bing_api_key",
				"BING_CONFIG_KEY": "env_bing_config_key",
				"OPENAI_API_KEY":  "env_openai_api_key",
			},
			expected: search.APIKeys{
				GoogleAPIKey:  "env_google_api_key",
				GoogleCSEID:   "env_google_cse_id",
				BingAPIKey:    "env_bing_api_key",
				BingConfigKey: "env_bing_config_key",
				OpenAIKey:     "env_openai_api_key",
			},
		},
		{
			name: "All keys from config files",
			files: map[string]string{
				"google-api-key":  "file_google_api_key",
				"google-cse-id":   "file_google_cse_id",
				"bing-api-key":    "file_bing_api_key",
				"bing-config-key": "file_bing_config_key",
				"openai-api-key":  "file_openai_api_key",
			},
			expected: search.APIKeys{
				GoogleAPIKey:  "file_google_api_key",
				GoogleCSEID:   "file_google_cse_id",
				BingAPIKey:    "file_bing_api_key",
				BingConfigKey: "file_bing_config_key",
				OpenAIKey:     "file_openai_api_key",
			},
		},
		{
			name: "Mixed sources",
			envVars: map[string]string{
				"GOOGLE_API_KEY": "env_google_api_key",
				"BING_API_KEY":   "env_bing_api_key",
			},
			files: map[string]string{
				"google-cse-id":   "file_google_cse_id",
				"bing-config-key": "file_bing_config_key",
				"openai-api-key":  "file_openai_api_key",
			},
			expected: search.APIKeys{
				GoogleAPIKey:  "env_google_api_key",
				GoogleCSEID:   "file_google_cse_id",
				BingAPIKey:    "env_bing_api_key",
				BingConfigKey: "file_bing_config_key",
				OpenAIKey:     "file_openai_api_key",
			},
		},
		{
			name:     "No keys set",
			expected: search.APIKeys{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("GOOGLE_API_KEY")
			os.Unsetenv("GOOGLE_CSE_ID")
			os.Unsetenv("BING_API_KEY")
			os.Unsetenv("BING_CONFIG_KEY")
			os.Unsetenv("OPENAI_API_KEY")

			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			for k, v := range tc.files {
				fmt.Printf("askWebConfigDir: %s\n", askWebConfigDir)
				fmt.Printf("k: %s\n", k)
				err := os.WriteFile(filepath.Join(askWebConfigDir, k), []byte(v), 0644)
				if err != nil {
					t.Fatalf("Failed to write config file %s: %v", k, err)
				}
			}

			keys := SetupKeys(askWebConfigDir)

			if keys != tc.expected {
				t.Errorf("Expected %+v, got %+v", tc.expected, keys)
			}

			for k := range tc.files {
				os.Remove(filepath.Join(askWebConfigDir, k))
			}
		})
	}
}
