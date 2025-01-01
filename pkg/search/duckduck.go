package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// For parsing HTML results:
	"github.com/PuerkitoBio/goquery"
)

const DDGRegion = "wt-wt"

func DDGSearch(query string, maxResults int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	client := &http.Client{
		Timeout: time.Duration(MaxTimeoutSeconds) * time.Second,
	}

	baseURL := "https://html.duckduckgo.com/html/"
	params := url.Values{}
	params.Add("q", query)
	params.Add("kl", DDGRegion)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	f, err := os.CreateTemp("/tmp", "ddgsearch")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	if _, err := f.Write(body); err != nil {
		return nil, fmt.Errorf("failed to write to temp file: %v", err)
	}
	f.Close()

	results, err := extractDDGResults(string(body), maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to extract results: %v", err)
	}

	return results, nil
}

func extractDDGResults(htmlContent string, maxResults int) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	numResults := 0
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		if numResults >= maxResults {
			return
		}

		result := SearchResult{}

		result.URL, _ = s.Find(".result__url").Attr("href")
		result.Title = s.Find(".result__title").Text()
		result.Snippet = s.Find(".result__snippet").Text()

		results = append(results, result)

		numResults++
	})

	return results, nil
}
