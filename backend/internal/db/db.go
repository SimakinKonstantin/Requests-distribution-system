package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(ConnectionString string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Open: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}
