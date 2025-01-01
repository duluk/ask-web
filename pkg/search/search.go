package search

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

type APIKeys struct {
	GoogleAPIKey  string
	GoogleCSEID   string
	BingAPIKey    string
	BingConfigKey string
	OpenAIKey     string
}

const MaxTimeoutSeconds = 10
