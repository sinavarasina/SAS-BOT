package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		jid TEXT PRIMARY KEY,
		number TEXT,
		username TEXT,
		previlege TEXT
	);
	`
	_, err = db.Exec(schema)
	return db, err
}
