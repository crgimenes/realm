package sqlite

import (
	"realm/model"
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
		expire_at datetime not null,
		logged_in integer not null,
		oauth_provider text not null,
		user_name text not null,
		avatar_url text not null
	);`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) SaveSession(sessionID string, sd *model.SessionData) error {
	// insert ou update
	sqlStatement := `
	insert into session (
		session_id,			-- 1
		expire_at,			-- 2
		logged_in,			-- 3
		oauth_provider,		-- 4
		user_name,			-- 5
		avatar_url			-- 6
	) values (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	) on conflict(session_id) do update set
		expire_at = $2,
		logged_in = $3,
		oauth_provider = $4,
		user_name = $5,
		avatar_url = $6;`

	_, err := s.DB.Exec(sqlStatement,
		sessionID,        // 1
		sd.ExpireAt,      // 2
		sd.LoggedIn,      // 3
		sd.OAuthProvider, // 4
		sd.UserName,      // 5
		sd.AvatarURL)     // 6

	return err
}

func (s *Sqlite) DeleteSession(sessionID string) error {
	sqlStatement := `delete from session where session_id = $1;`

	_, err := s.DB.Exec(sqlStatement, sessionID)

	return err
}

func (s *Sqlite) GetSession(sessionID string) (*model.SessionData, error) {
	sqlStatement := `select * from session where session_id = $1;`

	var data model.SessionData
	err := s.DB.Get(&data, sqlStatement, sessionID)

	return &data, err
}

func (s *Sqlite) DeleteExpiredSessions() error {
	sqlStatement := `delete from session where expire_at < datetime('now');`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) DeleteAllSessions() error {
	sqlStatement := `delete from session;`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) LoadAllSessions() (map[string]model.SessionData, error) {

	var (
		sqlStatement   = `select * from session;`
		sessionDataMap = make(map[string]model.SessionData)
		sessionData    []model.SessionData
		err            error
	)
	err = s.DB.Select(&sessionData, sqlStatement)
	if err != nil {
		return nil, err
	}

	for _, sd := range sessionData {
		sessionDataMap[sd.SessionID] = sd
	}

	return sessionDataMap, nil
}
