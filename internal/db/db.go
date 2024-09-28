package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

type Database struct {
	db *sql.DB
}

type DatabaseConnectionParams struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// NewDatabase creates a new MySQL database connection
func NewDatabase(params DatabaseConnectionParams) (*Database, error) {
	// Connect to the database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", params.Username, params.Password, params.Host, params.Port, params.Database))
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Set the max connections and connection lifetime
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(2 * time.Minute)

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Transaction(f func(tx *Transaction) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	transactionWrapper := &Transaction{tx: tx}
	err = f(transactionWrapper)
	if err != nil {
		dbErr := tx.Rollback()
		if dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to rollback transaction")
		}

		return err
	}

	return tx.Commit()
}

func (d *Database) Handle() *sql.DB {
	return d.db
}
