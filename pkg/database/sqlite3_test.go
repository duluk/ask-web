package database

import (
	"os"
	"testing"

	"ask-web/pkg/search"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const dbPath = "./test.db"
const dbTable = "conversations_test"

func TestMain(m *testing.M) {
	code := m.Run()

	os.Remove(dbPath)

	os.Exit(code)
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
}

func TestSaveSearchResults(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	var results []search.SearchResult
	results = append(results, search.SearchResult{
		Title:   "title",
		URL:     "url",
		Snippet: "snippet",
	})

	err = db.SaveSearchResults("query", results, "summary")
	assert.Nil(t, err)

	db.Close()
	RemoveDB()
}

func TestClose(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
	RemoveDB()
}

func TestInsertConversationWithError(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Insert a conversation with invalid data and assert there are errors
	// TODO: This won't do much at this point because InsertConversation doesn't do
	// much validation of iput, and Go's type system won't let me enter an
	// invalid argument. InsertConversation should, however, do some
	// validation. For instance, there are restrictions about temperature - eg,
	// 0.123 is technically invalid.
	// err = db.InsertConversation("prompt", "response", "", 0.0)
	// assert.NotNil(t, err)

	db.Close()
	RemoveDB()
}

func RemoveDB() {
	os.Remove(dbPath)
}
