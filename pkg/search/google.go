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

func GoogleSearch(apiKey string, cseID string, query string, maxResults int, filter FilterFunc) ([]SearchResult, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("cx", cseID)
	q.Set("q", query)
	n := float32(maxResults) * ExtraResultsFactor
	// fmt.Printf("maxResults: %d, ExtraResultsFactor: %f, n: %f, int(n): %d\n", maxResults, ExtraResultsFactor, n, int(n))
	q.Set("num", fmt.Sprintf("%d", int(n)))
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

	filteredResults := make([]SearchResult, 0, maxResults)
	for _, item := range result.Items {
		r := SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
		}

		if filter == nil || filter(r) {
			filteredResults = append(filteredResults, r)
			if len(filteredResults) == maxResults {
				break
			}
		}
	}

	return filteredResults, nil
}
