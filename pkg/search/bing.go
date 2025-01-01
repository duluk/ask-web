package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL = "https://api.bing.microsoft.com/v7.0/custom/search"
)

type SearchResponse struct {
	WebPages struct {
		Value []SearchResult `json:"value"`
	} `json:"webPages"`
}

func BingSearch(apiKey string, configKey string, query string, maxResults int) ([]SearchResult, error) {
	client := &http.Client{
		Timeout: time.Duration(MaxTimeoutSeconds) * time.Second,
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("count", fmt.Sprintf("%d", maxResults))
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

	return searchResp.WebPages.Value, nil
}
