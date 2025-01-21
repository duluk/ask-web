package main

// TODO: don't use wikipedia for results; too many tokens

import (
	"fmt"
	"log"
	"os"

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

// TODO:
// 1. Add a flag to specify whether other search engines should be used
// 2. Add a flag to specify the number of search results to use per search engine?

func main() {
	opts, err := config.Initialize()
	if err != nil {
		logger.Fatal("Error initializing config:", err)
		os.Exit(1)
	}

	var query string
	if pflag.NArg() > 0 {
		query = pflag.Arg(0)
	}

	apiKeys := utils.SetupKeys(opts.ConfigDir)

	if opts.ShowAPIKeys {
		fmt.Println("<== API keys ==>")
		fmt.Println("Google API Key:", apiKeys.GoogleAPIKey)
		fmt.Println("Bing API Key:", apiKeys.BingAPIKey)
		fmt.Println("Bing Config Key:", apiKeys.BingConfigKey)
		fmt.Println("CSE ID:", apiKeys.GoogleCSEID)
		fmt.Println("OpenAI Key:", apiKeys.OpenAIKey)
	}

	db, err := database.InitializeDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		logger.Fatal("Error opening database: ", err)
	}
	defer db.Close()

	var googleResults []search.SearchResult
	if apiKeys.GoogleAPIKey != "" && apiKeys.GoogleCSEID != "" {
		googleResults, err = search.GoogleSearch(apiKeys.GoogleAPIKey, apiKeys.GoogleCSEID, query, opts.NumResults)
		if err != nil {
			logger.Fatal("Error during web search:", err)
		}
	}
	logger.Info("Google Results:")
	for _, result := range googleResults {
		logger.Info(fmt.Sprintf("\t%s\n", result.URL))
	}

	var bingResults []search.SearchResult
	if apiKeys.BingAPIKey != "" && apiKeys.BingConfigKey != "" {
		bingResults, err = search.BingSearch(apiKeys.BingAPIKey, apiKeys.BingConfigKey, query, opts.NumResults)
		if err != nil {
			logger.Fatal("Error during web search:", err)
		}
	}
	logger.Info("Bing Results:")
	for _, result := range bingResults {
		logger.Info(fmt.Sprintf("\t%s\n", result.URL))
	}

	results := append(googleResults, ddgResults...)
	results = append(results, bingResults...)
	results = utils.DedupeResults(results)
	logger.Info("Final Results:")
	for _, result := range results {
		logger.Info(result.URL)
	}

	var contents []string
	for _, result := range results {
		logger.Info("Downloading:", result.URL)
		content, err := download.Page(result.URL)
		if err != nil {
			logger.Error("Error downloading page: ", err.Error())
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
	summary, err := summarize.Summarize(opts, apiKeys.OpenAIKey, cleanedContents, query, noOpClient)
	if err != nil {
		logger.Fatal("Error during summarization:", err)
	}

	logger.Info("Saving search results to database...")
	db.SaveSearchResults(query, results, summary)

	wrapper := linewrap.NewLineWrapper(80, 4, os.Stdout)
	wrapper.Write([]byte(summary))
	fmt.Println()
}
