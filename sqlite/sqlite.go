package sqlite

import (
	"realm/model"
	"realm/util"
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

func (s *Sqlite) CreateSessionTables() error {
	sqlStatement := `
	create table if not exists session (
		session_id text primary key,
		user_id text not null,
		expire_at datetime not null,
		logged_in integer not null,
		oauth_provider text not null,
		oauth_user_id text not null,
		user_name text not null,
		avatar_url text not null
	);`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) CreateUserTables() error {
	sqlStatement := `
	create table if not exists user (
		user_id text primary key,
		oauth_provider text not null,
		oauth_user_id text not null,
		user_name text not null,
		avatar_url text not null
	);`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) SaveUser(user *model.User) (model.User, error) {
	// insert ou update (on conflict) e retorna o user

	if user.ID == "" {
		// insert
		sqlStatement := `
		insert into user (
			user_id,
			oauth_provider,
			oauth_user_id,
			user_name,
			avatar_url
		) values (
			$1,
			$2,
			$3,
			$4,
			$5
		);`

		user.ID = util.RandomID()
		_, err := s.DB.Exec(sqlStatement,
			user.ID,
			user.OAuthProvider,
			user.OAuthUserID,
			user.UserName,
			user.AvatarURL)

		return *user, err
	}

	// update
	sqlStatement := `
	update user set
		user_name = $1,
		avatar_url = $2
	where id = $3;`

	_, err := s.DB.Exec(sqlStatement,
		user.UserName,
		user.AvatarURL,
		user.ID)

	return *user, err
}

func (s *Sqlite) GetUserFromOAuthID(oauthProvider string, oauthUserID string) (*model.User, error) {
	sqlStatement := `select
		id,
		oauth_provider,
		oauth_user_id,
		user_name,
		avatar_url
	from user 
	where oauth_provider = $1 
	and oauth_user_id = $2;`

	var user model.User
	err := s.DB.Get(&user,
		sqlStatement,
		oauthProvider, // 1
		oauthUserID)   // 2

	return &user, err
}

func (s *Sqlite) SaveSession(sessionID string, sd *model.SessionData) error {
	// insert ou update
	sqlStatement := `
	insert into session (
		session_id,			-- 1
		user_id,			-- 2
		expire_at,			-- 3
		logged_in,			-- 4
		oauth_provider,		-- 5
		oauth_user_id,		-- 6
		user_name,			-- 7
		avatar_url			-- 8
	) values (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	) on conflict(session_id) do update set
		user_id = $2,
		expire_at = $3,
		logged_in = $4,
		oauth_provider = $5,
		oauth_user_id = $6,
		user_name = $7,
		avatar_url = $8;`

	_, err := s.DB.Exec(sqlStatement,
		sessionID,        // 1
		sd.UserID,        // 2
		sd.ExpireAt,      // 3
		sd.LoggedIn,      // 4
		sd.OAuthProvider, // 5
		sd.OAuthUserID,   // 6
		sd.UserName,      // 7
		sd.AvatarURL)     // 8

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
		sqlStatement   = `select * from session where expire_at > datetime('now');`
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
