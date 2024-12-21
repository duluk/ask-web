package utils

import (
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
