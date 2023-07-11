package sqlite

import (
	"realm/session"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Sqlite struct {
	DB *sqlx.DB
}

var DB Sqlite

func Open(filename string) error {
	databaseName := strings.Join([]string{
		"file:",
		filename,
		"?cache=shared&mode=rwc"}, "")

	db, err := sqlx.Connect("sqlite", databaseName)
	if err != nil {
		return err
	}

	DB = Sqlite{
		DB: db,
	}

	return nil
}

func (s *Sqlite) Close() error {
	return s.DB.Close()
}

func (s *Sqlite) CreateTables() error {
	return nil
}

func (s *Sqlite) SaveSession(sessionID string, sd *session.SessionData) error {
	return nil
}

func (s *Sqlite) DeleteSession(sessionID string) error {
	return nil
}

func (s *Sqlite) GetSession(sessionID string) (*session.SessionData, error) {
	return nil, nil
}
