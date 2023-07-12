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
	sqlStatement := `
	create table if not exists session (
		session_id text primary key,
		expire_at integer not null,
		data text not null
	);`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) SaveSession(sessionID string, sd *session.SessionData) error {
	// insert ou update
	sqlStatement := `
	insert into session (
		session_id, 
		expire_at, 
		data) 
	values ($1, $2, $3)
	on conflict(session_id) 
	do update set expire_at = $2, data = $3;`

	_, err := s.DB.Exec(sqlStatement, sessionID, sd.ExpireAt, sd.Data)

	return err
}

func (s *Sqlite) DeleteSession(sessionID string) error {
	sqlStatement := `
	delete from session where session_id = $1;`

	_, err := s.DB.Exec(sqlStatement, sessionID)

	return err
}

func (s *Sqlite) GetSession(sessionID string) (*session.SessionData, error) {
	sqlStatement := `
	select data from session where session_id = $1;`

	var data session.SessionData
	err := s.DB.Get(&data, sqlStatement, sessionID)

	return &data, err
}
