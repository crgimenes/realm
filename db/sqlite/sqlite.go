package sqlite

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Sqlite struct {
	db *sqlx.DB
}

func New(filename string) (*Sqlite, error) {
	databaseName := filename + ".db"
	db, err := sqlx.Connect("sqlite", databaseName)
	if err != nil {
		return nil, err
	}

	s := &Sqlite{
		db: db,
	}
	return s, err
}
