package summarize

import (
	"fmt"

	"ask-web/pkg/config"
)

const (
	ModelOpenAI = "chatgpt"
	ModelGoogle = "gemini"
)

func NewSummarizer(model string, apiKey string, opts *config.Opts) (Summarizer, error) {
	switch model {
	case ModelOpenAI:
		return NewOpenAISummarizer(apiKey, opts)
	case ModelGoogle:
		return NewGoogleSummarizer(apiKey, opts)
	default:
		return nil, fmt.Errorf("unsupported model: %s. Supported models: %s, %s",
			model, ModelOpenAI, ModelGoogle)
	}
}
