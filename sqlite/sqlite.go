package sqlite

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNoRows = sql.ErrNoRows
)

type DB struct {
	db *sqlx.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", path)
	return &DB{db}, err
}

// Get gets a single entry from the database in the format of out interface
func (d *DB) Get(query string, in map[string]interface{}, out interface{}) error {
	boundQuery, args, err := d.db.BindNamed(query, in)
	if err != nil {
		return err
	}
	return d.db.Get(out, boundQuery, args...)
}

// Select gets multiple entries from the database in the format of out interface (use a slice [])
func (d *DB) Select(query string, in map[string]interface{}, out interface{}) error {
	boundQuery, args, err := d.db.BindNamed(query, in)
	if err != nil {
		return err
	}
	return d.db.Select(out, boundQuery, args...)
}

// NamedExec executes a named exec on our transaction
func (d *DB) NamedExec(command string, in interface{}) (int64, error) {
	id := int64(0)
	result, err := d.db.NamedExec(command, in)
	if result != nil {
		id, _ = result.LastInsertId()
	}
	return id, err
}

// Delete should be used for deletes
func (d *DB) Delete(command string, in interface{}) error {
	_, err := d.db.NamedExec(command, in)
	return err
}
