package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"ask-web/pkg/config"
	"ask-web/pkg/database"
	"ask-web/pkg/download"
	"ask-web/pkg/linewrap"
	"ask-web/pkg/logger"
	"ask-web/pkg/search"
	"ask-web/pkg/summarize"
	"ask-web/pkg/utils"
)

func main() {
	opts, err := config.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %s", err)
		os.Exit(1)
	}

	err = logger.Init(opts)
	if err != nil {
		panic(err)
	}
	log := logger.GetLogger()

	if opts.DumpConfig {
		config.DumpConfig(opts)
		os.Exit(0)
	}

	resultFilter := func(result search.SearchResult) bool {
		for _, url := range opts.FilteredURLs {
			if strings.Contains(result.URL, url) {
				return false
			}
		}

		return true
	}

	apiKeys := utils.SetupKeys(opts.ConfigDir)

	if opts.ShowAPIKeys {
		fmt.Println("<== API keys ==>")
		fmt.Println("Gemini API Key:", apiKeys.GeminiAPIKey)
		fmt.Println("Google API Key:", apiKeys.GoogleAPIKey)
		fmt.Println("Bing API Key:", apiKeys.BingAPIKey)
		fmt.Println("Bing Config Key:", apiKeys.BingConfigKey)
		fmt.Println("CSE ID:", apiKeys.GoogleCSEID)
		fmt.Println("OpenAI Key:", apiKeys.OpenAIKey)
	}

	var query string
	if pflag.NArg() > 0 {
		query, err = search.CreateSearchQuery(opts, apiKeys.OpenAIKey, pflag.Arg(0))
		if err != nil {
			query = pflag.Arg(0)
		}

		query = strings.Trim(query, "\"")
		query = url.QueryEscape(query)

		log.Info("Original prompt: ", pflag.Arg(0))
		log.Info("Generated query: ", query)
	}

	db, err := database.InitializeDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	defer db.Close()

	fmt.Println("Gathering search results for query:", query)
	var ddgResults []search.SearchResult
	ddgResults, err = search.DDGSearch(query, opts.NumResults, resultFilter)
	if err != nil {
		log.Fatal("Error during web search:", err)
	}
	for _, result := range ddgResults {
		log.Info("DuckDuckGo URL:", result.URL)
	}

	var googleResults []search.SearchResult
	if apiKeys.GoogleAPIKey != "" && apiKeys.GoogleCSEID != "" {
		googleResults, err = search.GoogleSearch(apiKeys.GoogleAPIKey, apiKeys.GoogleCSEID, query, opts.NumResults, resultFilter)
		if err != nil {
			log.Fatal("Error during web search:", err)
		}
	}
	for _, result := range googleResults {
		log.Info("Google URL:", result.URL)
	}

	var bingResults []search.SearchResult
	if apiKeys.BingAPIKey != "" && apiKeys.BingConfigKey != "" {
		bingResults, err = search.BingSearch(apiKeys.BingAPIKey, apiKeys.BingConfigKey, query, opts.NumResults, resultFilter)
		if err != nil {
			log.Fatal("Error during web search:", err)
		}
	}
	for _, result := range bingResults {
		log.Info("Bing URL:", result.URL)
	}

	results := append(ddgResults, googleResults...)
	results = append(results, bingResults...)
	results = utils.DedupeResults(results)

	fmt.Println("Downloading search results...")
	var contents []string
	for _, result := range results {
		log.Info("Downloading unique URL:", result.URL)
		content, err := download.Page(result.URL)
		if err != nil {
			log.Error(fmt.Sprintf("Error downloading %s: %s", result.URL, err.Error()))
			continue
		}
		contents = append(contents, content)
	}

	var cleanedContents []string
	for _, content := range contents {
		cleanedContents = append(cleanedContents, utils.CleanText(content))
	}

	// Set noOpClient to nil to use the real OpenAI API; the mock client is
	// used for testing. I'm not sure I like this method.
	var noOpClient summarize.OpenAIClient
	fmt.Println("Summarizing content...")

	summarizer := summarize.NewOpenAISummarizer(apiKeys.OpenAIKey, opts, noOpClient)
	summary, err := summarizer.Summarize(context.Background(), cleanedContents, query)
	if err != nil {
		log.Fatal("Error during summarization:", err)
	}

	db.SaveSearchResults(query, results, summary)

	wrapper := linewrap.NewLineWrapper(80, 4, os.Stdout)
	wrapper.Write([]byte(summary))
	fmt.Println()
}
