package database

// Use this module like this:
// db := NewDB("path/to/database.db")
// SaveSearchResults("query", searchResults, "summary")

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"ask-web/pkg/search"
	_ "github.com/mattn/go-sqlite3"
)

type SearchDB struct {
	db      *sql.DB
	dbTable string
}

// Retun errors to the caller in case we want to ignore them. That is, just
// because we can't store the conversations in the database doesn't mean we
// should stop the program.
func NewDB(dbPath string, dbTable string) (*SearchDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	_, err = db.Exec(DBSchema(dbTable))
	if err != nil {
		return nil, fmt.Errorf("error creating %s table: %v", dbTable, err)
	}

	sqlDB := SearchDB{}
	sqlDB.db = db
	sqlDB.dbTable = dbTable
	return &sqlDB, nil
}

func (sqlDB *SearchDB) SaveSearchResults(query string, results []search.SearchResult, summary string) error {
	// extract URLs from search results
	var urls []string
	for _, result := range results {
		urls = append(urls, result.URL)
	}
	urlsJSON, err := json.Marshal(urls)
	if err != nil {
		panic(err)
	}

	stmt, err := sqlDB.db.Prepare(`
	INSERT INTO ` + sqlDB.dbTable + `(query, results, summary)
	VALUES(?, ?, ?)
	`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(query, urlsJSON, summary)
	if err != nil {
		panic(err)
	}

	return nil
}

func (sqlDB *SearchDB) Close() {
	err := sqlDB.db.Close()
	if err != nil {
		log.Fatalf("error closing database: %v", err)
	}
}
