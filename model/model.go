package model

import "time"

type SessionData struct {
	ExpireAt      time.Time `db:"expire_at"`
	LoggedIn      bool      `db:"logged_in"`
	OAuthProvider string    `db:"oauth_provider"`
	UserName      string    `db:"user_name"`
	AvatarURL     string    `db:"avatar_url"`
	SessionID     string    `db:"session_id"`
}
