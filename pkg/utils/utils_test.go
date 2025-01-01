package utils

import (
	"os"
	"reflect"
	"testing"

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
	// Backup the original environment variables
	originalEnv := make(map[string]string)
	theKeys := []string{"GOOGLE_API_KEY", "GOOGLE_CSE_ID", "BING_API_KEY", "BING_CONFIG_KEY", "OPENAI_API_KEY"}
	for _, key := range theKeys {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		// Restore environment variables after the test
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}()

	// Set up the environment variables for testing
	os.Setenv("GOOGLE_API_KEY", "test-google-api-key")
	os.Setenv("GOOGLE_CSE_ID", "test-google-cse-id")
	os.Setenv("BING_API_KEY", "test-bing-api-key")
	os.Setenv("BING_CONFIG_KEY", "test-bing-config-key")
	os.Setenv("OPENAI_API_KEY", "test-openai-api-key")

	keys := SetupKeys()

	expected := search.APIKeys{
		GoogleAPIKey:  "test-google-api-key",
		GoogleCSEID:   "test-google-cse-id",
		BingAPIKey:    "test-bing-api-key",
		BingConfigKey: "test-bing-config-key",
		OpenAIKey:     "test-openai-api-key",
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("SetupKeys() = %v; want %v", keys, expected)
	}

	// Test case where no environment variables are set
	for _, key := range []string{"GOOGLE_API_KEY", "GOOGLE_CSE_ID",
		"BING_API_KEY", "BING_CONFIG_KEY", "OPENAI_API_KEY"} {
		os.Unsetenv(key)
	}
	keys = SetupKeys()
	expected = search.APIKeys{
		GoogleAPIKey:  "",
		GoogleCSEID:   "",
		BingAPIKey:    "",
		BingConfigKey: "",
		OpenAIKey:     "",
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("SetupKeys() with no env vars = %v; want %v", keys, expected)
	}

	// Test case where keys are read from files
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME environment variable not set, skipping file based key test")
	}

	// Create dummy files for the test
	os.MkdirAll(home+"/.config/ask-web/", 0755)
	for _, key := range []string{"google-api-key", "google-cse-id", "bing-api-key", "bing-config-key", "openai-api-key"} {
		file, err := os.Create(home + "/.config/ask-web/" + key)
		if err != nil {
			t.Fatalf("Failed to create test key file %s: %v", key, err)
		}
		_, err = file.WriteString("file-" + key)
		if err != nil {
			t.Fatalf("Failed to write to test key file %s: %v", key, err)
		}
		file.Close()
	}

	keys = SetupKeys()
	expected = search.APIKeys{
		GoogleAPIKey:  "file-google-api-key",
		GoogleCSEID:   "file-google-cse-id",
		BingAPIKey:    "file-bing-api-key",
		BingConfigKey: "file-bing-config-key",
		OpenAIKey:     "file-openai-api-key",
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("SetupKeys() with file based keys = %v; want %v", keys,
			expected)
	}

	// cleanup dummy files
	for _, key := range []string{"google-api-key", "google-cse-id", "bing-api-key", "bing-config-key", "openai-api-key"} {
		os.Remove(home + "/.config/ask-web/" + key)
	}
	os.Remove(home + "/.config/ask-web")

}
