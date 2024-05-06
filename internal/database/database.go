package database

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbInstance *sql.DB

func New() (*sql.DB, error) {
	// reuse connection
	if dbInstance != nil {
		return dbInstance, nil
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, errors.New("database url not found")
	}

	// sql.Open() does not create connection
	// all it does is initialize the pool for future use
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	// to verify that everything is set up correctly we need to use the
	// db.Ping() method to create a connection and check for any errors
	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbInstance = db

	return dbInstance, nil
}
