package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"ask-web/pkg/download"
	"ask-web/pkg/linewrap"
	"ask-web/pkg/search"
	"ask-web/pkg/summarize"
	"ask-web/pkg/utils"
)

func main() {
	numResults := flag.Int("n", 3, "Number of search results to use")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Error: Search query is required.")
		os.Exit(1)
	}
	query := flag.Arg(0)

	fmt.Println("Searching for:", query)
	results, err := search.GoogleSearch(query, *numResults)
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
	summary, err := summarize.Summarize(cleanedContents, query)
	if err != nil {
		log.Fatal("Error during summarization:", err)
	}

	wrapper := linewrap.NewLineWrapper(80, 4, os.Stdout)
	wrapper.Write([]byte(summary))
	fmt.Println()
}
