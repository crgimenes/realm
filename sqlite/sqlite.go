package sqlite

import (
	"database/sql"
	"log"
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
		id text primary key,
		oauth_provider text not null,
		oauth_user_id text not null,
		user_name text not null,
		avatar_url text not null
	);`

	_, err := s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) SaveUser(user *model.User) (model.User, error) {

	u, err := s.GetUserFromOAuthID(user.OAuthProvider, user.OAuthUserID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
			return *user, err
		}
	}

	user.ID = u.ID

	if user.ID == "" {
		// insert
		sqlStatement := `
		insert into user (
			id,
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

	_, err = s.DB.Exec(sqlStatement,
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

/////////////////////////////////////////////////////////////////
// forum

func (s *Sqlite) CreateForumTables() error {
	sqlStatement := `
	create table if not exists forum (
		name text not null,
		name_slug text not null,
		primary key(name_slug)
	);`

	_, err := s.DB.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `
	create table if not exists thread (
		id text not null,
		forum_name text not null,
		title text not null,
		content text not null,
		user_id text not null,
		created_at datetime not null,
		updated_at datetime not null,
		primary key(id),
		foreign key(forum_name) references forum(name_slug)
	);`

	_, err = s.DB.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `
	create table if not exists comment (
		id text not null,
		thread_id text not null,
		user_id text not null,
		content text not null,
		created_at datetime not null,
		updated_at datetime not null,
		primary key(id),
		foreign key(thread_id) references thread(id)
	);`

	_, err = s.DB.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return err
}

func (s *Sqlite) CreateForum(name string) error {
	sqlStatement := `
	insert into forum (
		name,
		name_slug
	) values (
		$1,
		$2
	);`

	_, err := s.DB.Exec(sqlStatement,
		name,
		strings.ToLower(name))

	return err
}

func (s *Sqlite) GetForum(name string) (*model.Forum, error) {
	sqlStatement := `select * from forum where name_slug = $1;`

	var forum model.Forum
	err := s.DB.Get(&forum, sqlStatement, strings.ToLower(name))

	return &forum, err
}

func (s *Sqlite) GetForumList() ([]model.Forum, error) {
	sqlStatement := `select * from forum;`

	var forumList []model.Forum
	err := s.DB.Select(&forumList, sqlStatement)

	return forumList, err
}

func (s *Sqlite) DeleteForum(name string) error {
	sqlStatement := `delete from forum where name_slug = $1;`

	_, err := s.DB.Exec(sqlStatement, strings.ToLower(name))

	return err
}

func (s *Sqlite) CreateThread(thread *model.Thread) error {
	sqlStatement := `
	insert into thread (
		id,
		forum_name,
		title,
		content,
		user_id,
		created_at,
		updated_at
	) values (
		$1,
		$2,
		$3,
		$4,
		$5,
		datetime('now'),
		datetime('now')
	);`

	_, err := s.DB.Exec(sqlStatement,
		thread.ID,
		thread.ForumName,
		thread.Title,
		thread.Content,
		thread.UserID)

	return err
}

func (s *Sqlite) GetThread(id string) (*model.Thread, error) {
	sqlStatement := `select * from thread where id = $1;`

	var thread model.Thread
	err := s.DB.Get(&thread, sqlStatement, id)

	return &thread, err
}

func (s *Sqlite) GetThreadList(forumName string) ([]model.Thread, error) {
	sqlStatement := `select * from thread where forum_name = $1;`

	var threadList []model.Thread
	err := s.DB.Select(&threadList, sqlStatement, forumName)

	return threadList, err
}

func (s *Sqlite) DeleteThread(id string) error {
	sqlStatement := `delete from thread where id = $1;`

	_, err := s.DB.Exec(sqlStatement, id)

	return err
}

func (s *Sqlite) CreateComment(comment *model.Comment) error {
	sqlStatement := `
	insert into comment (
		id,
		thread_id,
		user_id,
		content,
		created_at,
		updated_at
	) values (
		$1,
		$2,
		$3,
		$4,
		datetime('now'),
		datetime('now')
	);`

	_, err := s.DB.Exec(sqlStatement,
		comment.ID,
		comment.ThreadID,
		comment.UserID,
		comment.Content)

	return err
}

func (s *Sqlite) GetComment(id string) (*model.Comment, error) {
	sqlStatement := `select * from comment where id = $1;`

	var comment model.Comment
	err := s.DB.Get(&comment, sqlStatement, id)

	return &comment, err
}

func (s *Sqlite) GetCommentList(threadID string) ([]model.Comment, error) {
	sqlStatement := `select * from comment where thread_id = $1;`

	var commentList []model.Comment
	err := s.DB.Select(&commentList, sqlStatement, threadID)

	return commentList, err
}

func (s *Sqlite) DeleteComment(id string) error {
	sqlStatement := `delete from comment where id = $1;`

	_, err := s.DB.Exec(sqlStatement, id)

	return err
}

/////////////////////////////////////////////////////////////////
// chat

func (s *Sqlite) CreateChatTables() error {
	sqlStatement := `
	create table if not exists chat_room (
		name text not null,
		name_slug text not null,
		created_at datetime not null,
		updated_at datetime not null,
		primary key(name_slug)
	);`

	_, err := s.DB.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `
	create table if not exists chat_message (
		id text not null,
		room_id text not null,
		user_id text not null,
		content text not null,
		created_at datetime not null,
		updated_at datetime not null,
		primary key(id),
		foreign key(room_id) references chat_room(name_slug)
	);`

	_, err = s.DB.Exec(sqlStatement)

	return err
}

func (s *Sqlite) CreateChatRoom(name string) error {
	sqlStatement := `
	insert into chat_room (
		name,
		name_slug,
		created_at,
		updated_at
	) values (
		$1,
		$2,
		datetime('now'),
		datetime('now')
	);`

	_, err := s.DB.Exec(sqlStatement,
		name,
		strings.ToLower(name))

	return err
}

func (s *Sqlite) GetChatRoom(name string) (*model.ChatRoom, error) {
	sqlStatement := `select * from chat_room where name_slug = $1;`

	var chatRoom model.ChatRoom
	err := s.DB.Get(&chatRoom, sqlStatement, strings.ToLower(name))

	return &chatRoom, err
}

func (s *Sqlite) GetChatRoomList() ([]model.ChatRoom, error) {
	sqlStatement := `select * from chat_room;`

	var chatRoomList []model.ChatRoom
	err := s.DB.Select(&chatRoomList, sqlStatement)

	return chatRoomList, err
}

func (s *Sqlite) DeleteChatRoom(name string) error {
	sqlStatement := `delete from chat_room where name_slug = $1;`

	_, err := s.DB.Exec(sqlStatement, strings.ToLower(name))

	return err
}

func (s *Sqlite) CreateChatMessage(message *model.ChatMessage) error {
	sqlStatement := `
	insert into chat_message (
		id,
		room_id,
		user_id,
		content,
		created_at,
		updated_at
	) values (
		$1,
		$2,
		$3,
		$4,
		datetime('now'),
		datetime('now')
	);`

	_, err := s.DB.Exec(sqlStatement,
		message.ID,
		message.RoomID,
		message.UserID,
		message.Content)

	return err
}

func (s *Sqlite) GetChatMessage(id string) (*model.ChatMessage, error) {
	sqlStatement := `select * from chat_message where id = $1;`

	var chatMessage model.ChatMessage
	err := s.DB.Get(&chatMessage, sqlStatement, id)

	return &chatMessage, err
}

func (s *Sqlite) GetChatMessageList(roomID string) ([]model.ChatMessage, error) {
	sqlStatement := `select * from chat_message where room_id = $1;`

	var chatMessageList []model.ChatMessage
	err := s.DB.Select(&chatMessageList, sqlStatement, roomID)

	return chatMessageList, err
}

func (s *Sqlite) DeleteChatMessage(id string) error {
	sqlStatement := `delete from chat_message where id = $1;`

	_, err := s.DB.Exec(sqlStatement, id)

	return err
}

/////////////////////////////////////////////////////////////////
