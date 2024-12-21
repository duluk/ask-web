package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

type googleSearchResult struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

func GoogleSearch(query string, numResults int) ([]SearchResult, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	cseID := os.Getenv("GOOGLE_CSE_ID")

	if apiKey == "" || cseID == "" {
		return nil, errors.New("environment variables GOOGLE_API_KEY and GOOGLE_CSE_ID must be set")
	}

	baseURL := "https://www.googleapis.com/customsearch/v1"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", apiKey)
	q.Set("cx", cseID)
	q.Set("q", query)
	q.Set("num", fmt.Sprintf("%d", numResults))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google Search API error: %s", resp.Status)
	}

	var result googleSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var searchResults []SearchResult
	for _, item := range result.Items {
		searchResults = append(searchResults, SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}

	return searchResults, nil
}
