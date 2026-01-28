package dbpkg

import (
	"database/sql"
	"log"
)

func CreateTables(db *sql.DB) error {
	_, err := db.Query(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		first_name TEXT,
		last_name TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)

	if err != nil {
		return err
	} else {
		log.Println("Users table created")
	}

	return nil
}
