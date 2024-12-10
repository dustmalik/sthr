package db

import (
	"database/sql"
	"log"
	"net/url"

	"github.com/MunifTanjim/stremthru/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	URI ConnectionURI
}

var db *DB
var dialect string

var CurrentTimestamp string

func Exec(query string, args ...any) (sql.Result, error) {
	return db.Exec(adaptQuery(query), args...)
}

func Query(query string, args ...any) (*sql.Rows, error) {
	return db.Query(adaptQuery(query), args...)
}

func QueryRow(query string, args ...any) *sql.Row {
	return db.QueryRow(adaptQuery(query), args...)
}

type Tx struct {
	tx *sql.Tx
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return tx.tx.Exec(adaptQuery(query), args...)
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

func Begin() (*Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func Ping() {
	err := db.Ping()
	if err != nil {
		log.Fatalf("[db] failed to ping: %v\n", err)
	}
	one := 0
	row := QueryRow("SELECT 1")
	if err := row.Scan(&one); err != nil {
		log.Fatalf("[db] failed to query: %v\n", err)
	}
}

func Open() *DB {
	uri, err := ParseConnectionURI(config.DatabaseURI)
	if err != nil {
		log.Fatalf("[db] failed to parse uri: %v\n", err)
	}

	dialect = uri.Dialect
	dsnModifiers := []DSNModifier{}

	switch dialect {
	case "sqlite":
		CurrentTimestamp = "unixepoch()"
		dsnModifiers = append(dsnModifiers, func(u *url.URL, q *url.Values) {
			u.Scheme = "file"
		})
	case "postgres":
		CurrentTimestamp = "current_timestamp"
	default:
		log.Fatalf("[db] unsupported dialect: %v\n", dialect)
	}

	database, err := sql.Open(uri.driverName, uri.DSN(dsnModifiers...))
	if err != nil {
		log.Fatalf("[db] failed to open: %v\n", err)
	}
	db = &DB{
		DB:  database,
		URI: uri,
	}

	return db
}

func Close() error {
	return db.Close()
}
