package model

import "time"

type SessionData struct {
	UserID        string    `db:"user_id"`
	ExpireAt      time.Time `db:"expire_at"`
	LoggedIn      bool      `db:"logged_in"`
	OAuthProvider string    `db:"oauth_provider"`
	OAuthUserID   string    `db:"oauth_user_id"`
	UserName      string    `db:"user_name"`
	AvatarURL     string    `db:"avatar_url"`
	SessionID     string    `db:"session_id"`
}

type User struct {
	ID            string `db:"id"`
	OAuthProvider string `db:"oauth_provider"`
	OAuthUserID   string `db:"oauth_user_id"`
	UserName      string `db:"user_name"`
	AvatarURL     string `db:"avatar_url"`
}

// forum

type Forum struct {
	Name     string `db:"name"`
	NameSlug string `db:"name_slug"`
}

type Thread struct {
	ID        string    `db:"id"`
	ForumName string    `db:"forum_name"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Comment struct {
	ID        string    `db:"id"`
	ThreadID  string    `db:"thread_id"`
	UserID    string    `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// chat

type ChatRoom struct {
	Name      string    `db:"name"`
	NameSlug  string    `db:"name_slug"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ChatMessage struct {
	ID        string    `db:"id"`
	RoomID    string    `db:"room_id"`
	UserID    string    `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Category struct {
	NameSlug string `db:"name_slug"`
	Name     string `db:"name"`
}
