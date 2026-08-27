package postgres

import (
	"database/sql"

	// Register the pgx driver under the "pgx" name for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a PostgreSQL connection pool and verifies connectivity.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
