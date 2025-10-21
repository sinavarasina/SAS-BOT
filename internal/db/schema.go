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
	CREATE TABLE IF NOT EXISTS pengaduan (
		id SERIAL PRIMARY KEY,
		user_jid TEXT,
		deskripsi TEXT,
		pict_path TEXT,
		sent_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	return db, err
}
