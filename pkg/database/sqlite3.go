package database

// Use this module like this:
// db := NewDB("path/to/database.db")
// db.SaveSearchResults("query", searchResults, "summary")

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"ask-web/pkg/search"
	_ "github.com/mattn/go-sqlite3"
)

type SearchRow struct {
	ID    int
	Query string
}

type ResultRow struct {
	Query   string
	Summary string
}

type SearchDB struct {
	db      *sql.DB
	dbTable string
}

// Retun errors to the caller in case we want to ignore them. That is, just
// because we can't store the conversations in the database doesn't mean we
// should stop the program.
func NewDB(dbPath string, dbTable string) (*SearchDB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %v", err)
		}
		file.Close()
	}

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

func (sqlDB *SearchDB) SearchForResult(keyword string) ([]SearchRow, error) {
	rows, err := sqlDB.db.Query(`
		SELECT id, query FROM `+sqlDB.dbTable+` WHERE summary LIKE ?;
	`, "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	defer rows.Close()

	var results []SearchRow
	for rows.Next() {
		var result SearchRow
		err := rows.Scan(&result.ID, &result.Query)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		// If response happens to be NULL (conv_id isn't set), it's fine to
		// just skip it
		if result.ID != 0 {
			results = append(results, SearchRow{ID: result.ID, Query: result.Query})
		}
	}

	return results, nil
}

func (sqlDB *SearchDB) ReturnSearchResult(sumID int) *ResultRow {
	rows, err := sqlDB.db.Query(`
		SELECT query, summary FROM `+sqlDB.dbTable+` WHERE id = ?;
	`, sumID)
	if err != nil {
		log.Fatalf("error showing conversation: %v", err)
	}
	defer rows.Close()

	var row ResultRow
	for rows.Next() {
		err := rows.Scan(&row.Query, &row.Summary)
		if err != nil {
			log.Fatalf("error showing conversation: %v", err)
		}

		return &row
	}

	return nil
}

func (sqlDB *SearchDB) ShowSearchResult(sumID int) {
	result := sqlDB.ReturnSearchResult(sumID)
	fmt.Printf("Prompt: %s\n", result.Query)
	fmt.Printf("Summary: %s\n", result.Summary)
}

func (sqlDB *SearchDB) Close() {
	err := sqlDB.db.Close()
	if err != nil {
		log.Fatalf("error closing database: %v", err)
	}
}
