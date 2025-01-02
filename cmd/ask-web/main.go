package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"ask-web/pkg/database"
	"ask-web/pkg/download"
	"ask-web/pkg/linewrap"
	"ask-web/pkg/search"
	"ask-web/pkg/summarize"
	"ask-web/pkg/utils"
)

type opts struct {
	DBFileName string
	DBTable    string
	numResults int
	numTokens  int
}

// TODO:
// 1. Add a flag to specify whether other search engines should be used
// 2. Add a flag to specify the number of search results to use per search engine?

func main() {
	configDir, err := xdg.ConfigFile("ask-web")
	if err != nil {
		log.Fatal("Error getting config directory:", err)
		os.Exit(1)
	}

	defaultDBFileName := filepath.Join(configDir, "search.db")

	dbFileName := flag.String("db", defaultDBFileName, "Database file name")
	dbTable := flag.String("table", "search_results", "Database table name")
	numResults := flag.Int("n", 3, "Number of search results to use")
	numTokens := flag.Int("t", 420, "Number of tokens to use for summarization")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Error: Search query is required.")
		os.Exit(1)
	}
	query := flag.Arg(0)

	apiKeys := utils.SetupKeys()

	// TODO: turn this into a flag
	// Display keys
	// fmt.Println("Google API Key:", apiKeys.GoogleAPIKey)
	// fmt.Println("Bing API Key:", apiKeys.BingAPIKey)
	// fmt.Println("Bing Config Key:", apiKeys.BingConfigKey)
	// fmt.Println("CSE ID:", apiKeys.GoogleCSEID)
	// fmt.Println("OpenAI Key:", apiKeys.OpenAIKey)

	opts := opts{
		DBFileName: *dbFileName,
		DBTable:    *dbTable,
		numResults: *numResults,
		numTokens:  *numTokens,
	}

	// If DB exists, it just opens it; otherwise, it creates it first
	db, err := database.InitializeDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		fmt.Println("Error opening database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	var googleResults []search.SearchResult
	if apiKeys.GoogleAPIKey != "" && apiKeys.GoogleCSEID != "" {
		googleResults, err = search.GoogleSearch(apiKeys.GoogleAPIKey, apiKeys.GoogleCSEID, query, opts.numResults)
		if err != nil {
			log.Fatal("Error during web search:", err)
		}
	}

	var ddgResults []search.SearchResult
	ddgResults, err = search.DDGSearch(query, opts.numResults)
	if err != nil {
		log.Fatal("Error during web search:", err)
	}

	var bingResults []search.SearchResult
	if apiKeys.BingAPIKey != "" && apiKeys.BingConfigKey != "" {
		bingResults, err = search.BingSearch(apiKeys.BingAPIKey, apiKeys.BingConfigKey, query, opts.numResults)
		if err != nil {
			log.Fatal("Error during web search:", err)
		}
	}

	results := append(googleResults, ddgResults...)
	results = append(results, bingResults...)
	results = utils.DedupeResults(results)

	var contents []string
	for _, result := range results {
		fmt.Println("Downloading:", result.URL)
		content, err := download.Page(result.URL)
		if err != nil {
			log.Println("Error downloading", result.URL, ":", err)
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
	summary, err := summarize.Summarize(apiKeys.OpenAIKey, cleanedContents, query, opts.numTokens, noOpClient)
	if err != nil {
		log.Fatal("Error during summarization:", err)
	}

	fmt.Println("Saving search results to database...")
	db.SaveSearchResults(query, results, summary)

	wrapper := linewrap.NewLineWrapper(80, 4, os.Stdout)
	wrapper.Write([]byte(summary))
	fmt.Println()
}
