package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
}

func main() {
	numResults := flag.Int("n", 3, "Number of search results to use")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Error: Search query is required.")
		os.Exit(1)
	}
	query := flag.Arg(0)

	apiKey, cseID, openAIKey := utils.SetupKeys()

	opts := opts{
		DBFileName: "/Users/jab3/.config/ask-web/search.db",
		DBTable:    "search_results",
		numResults: *numResults,
	}

	// If DB exists, it just opens it; otherwise, it creates it first
	db, err := database.InitializeDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		fmt.Println("Error opening database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("Searching for:", query)
	results, err := search.GoogleSearch(apiKey, cseID, query, opts.numResults)
	if err != nil {
		log.Fatal("Error during web search:", err)
	}

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

	fmt.Println("Summarizing content...")
	summary, err := summarize.Summarize(openAIKey, cleanedContents, query)
	if err != nil {
		log.Fatal("Error during summarization:", err)
	}

	fmt.Println("Saving search results to database...")
	db.SaveSearchResults(query, results, summary)

	wrapper := linewrap.NewLineWrapper(80, 4, os.Stdout)
	wrapper.Write([]byte(summary))
	fmt.Println()
}
