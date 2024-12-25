package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// CleanText removes HTML tags and extra whitespace from the text
func CleanText(text string) string {
	p := bluemonday.StrictPolicy()
	text = p.Sanitize(text)

	text = strings.Join(strings.Fields(text), " ")

	return text
}

func SetupKeys() (string, string, string) {
	apiKey := getKey("GOOGLE_API_KEY")
	cseID := getKey("GOOGLE_CSE_ID")
	openAIKey := getKey("OPENAI_API_KEY")

	return apiKey, cseID, openAIKey
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
