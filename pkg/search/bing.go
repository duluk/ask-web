package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const BaseURL = "https://api.bing.microsoft.com/v7.0/custom/search"

type SearchResponse struct {
	WebPages struct {
		Value []SearchResult `json:"value"`
	} `json:"webPages"`
}

func BingSearch(apiKey string, configKey string, query string, maxResults int, filter FilterFunc) ([]SearchResult, error) {
	client := &http.Client{
		Timeout: time.Duration(MaxTimeoutSeconds) * time.Second,
	}

	params := url.Values{}
	params.Add("q", query)
	n := float32(maxResults) * ExtraResultsFactor
	// fmt.Printf("maxResults: %d, ExtraResultsFactor: %f, n: %f, int(n): %d\n", maxResults, ExtraResultsFactor, n, int(n))
	params.Add("count", fmt.Sprintf("%d", int(n)))
	params.Add("customConfig", configKey)
	params.Add("safeSearch", "Off")

	reqURL := fmt.Sprintf("%s?%s", BaseURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed with status %d: %s",
			resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	filteredResults := make([]SearchResult, 0, maxResults)
	for _, result := range searchResp.WebPages.Value {
		if filter == nil || filter(result) {
			filteredResults = append(filteredResults, result)
			if len(filteredResults) == maxResults {
				break
			}
		}
	}

	return filteredResults, nil
}
