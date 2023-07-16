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
