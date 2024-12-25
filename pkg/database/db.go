package database

import (
	"database/sql"
	"strconv"
)

const SchemaVersion = 3

func DBSchema(dbTable string) string {
	return `
	CREATE TABLE IF NOT EXISTS ` + dbTable + ` (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		query TEXT NOT NULL,
		results TEXT NOT NULL,
		summary TEXT NOT NULL
	);
	`
}

func SchemaQueryV1(dbTable string) string {
	return ""
}

// There's got to be a better way to do this
func getSchemaSQL(schemaVersion int, dbTable string) string {
	switch schemaVersion {
	case 1:
		return SchemaQueryV1(dbTable)
	// case 2:
	// 	return SchemaQueryV2(dbTable)
	// case 3:
	// 	return SchemaQueryV3(dbTable)
	default:
		return ""
	}
}

func applySchema(db *sql.DB, dbTable string, schemaVersion int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(getSchemaSQL(schemaVersion, dbTable))

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func setSchemaVersion(db *sql.DB, schemaVersion int) error {
	verStr := strconv.Itoa(schemaVersion)
	_, err := db.Exec(`PRAGMA user_version = ` + verStr)
	return err
}

func InitializeDB(dbPath string, dbTable string) (*SearchDB, error) {
	// DB created only if it doesn't exist
	searchDB, err := NewDB(dbPath, dbTable)
	if err != nil {
		return searchDB, err
	}

	var currentVersion int
	err = searchDB.db.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return searchDB, err
	}

	if currentVersion == 0 {
		// This should mean it's the first time we've created this database,
		// which means it should be using the latest schema, which should mean
		// the latest schema version. So just set that.
		setSchemaVersion(searchDB.db, SchemaVersion)
	} else if currentVersion < SchemaVersion {
		// If the current schema exists but is less than the latest schema,
		// apply each schema that was missed.
		for i := currentVersion + 1; i <= SchemaVersion; i++ {
			err = applySchema(searchDB.db, dbTable, i)
			if err != nil {
				return searchDB, err
			}
		}
	}

	return searchDB, err
}
