package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SQL_DRIVER = "sqlite3"
)

type SQLiteOptions struct {
	FILENAME string
}

type SQLiteClient struct {
	db *sql.DB
}

func NewSQLiteClient(opts SQLiteOptions) (*SQLiteClient, error) {
	db, err := sql.Open(SQL_DRIVER, opts.FILENAME)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &SQLiteClient{
		db: db,
	}, nil
}

func (client *SQLiteClient) Close() {
	client.db.Close()
}
