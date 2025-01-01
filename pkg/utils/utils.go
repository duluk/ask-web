package utils

import (
	"bufio"
	"os"
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
}

func getKey(keyUpper string) string {
	key := os.Getenv(keyUpper)

	// TODO this should attempt XDG_CONFIG_HOME first, then HOME
	if key == "" {
		home := os.Getenv("HOME")
		keyLower := strings.ToLower(keyUpper)
		keyLower = strings.ReplaceAll(keyLower, "_", "-")
		file, err := os.Open(home + "/.config/ask-web/" + keyLower)
		if err != nil {
			fmt.Printf("No ENV for %s and Error reading file: %s", keyUpper, err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			key = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("No ENV for %s and Error reading file: %s", keyUpper, err)
			os.Exit(1)
		}
	}

	return key
}
