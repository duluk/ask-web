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

type FilterFunc func(SearchResult) bool

const MaxTimeoutSeconds = 10
const ExtraResultsFactor = 2.0
