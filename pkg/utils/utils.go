package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"ask-web/pkg/search"
	"github.com/microcosm-cc/bluemonday"
)

// CleanText removes HTML tags and extra whitespace from the text
func CleanText(text string) string {
	p := bluemonday.StrictPolicy()
	text = p.Sanitize(text)

	text = strings.Join(strings.Fields(text), " ")

	return text
}

func DedupeResults(results []search.SearchResult) []search.SearchResult {
	seen := make(map[string]bool)
	dedupedResults := make([]search.SearchResult, 0)

	for _, result := range results {
		// remove query parameters; this means any URLs that differ only in
		// query parameters will use only the first one seen. That may or may
		// not be what we want.
		cleanURL := strings.Split(result.URL, "?")[0]
		if _, ok := seen[cleanURL]; !ok {
			seen[cleanURL] = true
			dedupedResults = append(dedupedResults, result)
		}
	}
	return dedupedResults
}

func SetupKeys(configDir string) search.APIKeys {
	return search.APIKeys{
		GeminiAPIKey:  getKey("GEMINI_API_KEY", configDir),
		GoogleAPIKey:  getKey("GOOGLE_API_KEY", configDir),
		GoogleCSEID:   getKey("GOOGLE_CSE_ID", configDir),
		BingAPIKey:    getKey("BING_API_KEY", configDir),
		BingConfigKey: getKey("BING_CONFIG_KEY", configDir),
		OpenAIKey:     getKey("OPENAI_API_KEY", configDir),
	}
}

// Return empty string if no key is found so that we can just check for the
// existence of a key to decide if we should use it that engine
func getKey(keyUpper string, configDir string) string {
	key := os.Getenv(keyUpper)

	if key == "" {
		keyLower := strings.ToLower(keyUpper)
		keyLower = strings.ReplaceAll(keyLower, "_", "-")

		file, err := os.Open(filepath.Join(configDir, keyLower))
		if err != nil {
			return ""
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			key = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			key = ""
		}
	}

	return key
}
