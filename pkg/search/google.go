package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type googleSearchResult struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

func GoogleSearch(apiKey string, cseID string, query string, numResults int) ([]SearchResult, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("cx", cseID)
	q.Set("q", query)
	q.Set("num", fmt.Sprintf("%d", numResults))
	u.RawQuery = q.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-goog-api-key", apiKey)

	resp, err := client.Do(req)
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
