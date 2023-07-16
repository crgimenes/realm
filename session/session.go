package session

import (
	"log"
	"net/http"
	"time"

	"realm/globalconst"
	"realm/model"
	"realm/sqlite"
	"realm/util"
)

var (
	SC *Control // Session Control
)

type Control struct {
	cookieName string
	DataMap    map[string]model.SessionData
}

func New(cookieName string) {
	dm, err := sqlite.DB.LoadAllSessions()
	if err != nil {
		panic(err)
	}

	SC = &Control{
		cookieName: cookieName,
		DataMap:    dm,
	}
}

func (c *Control) Get(r *http.Request) (string, *model.SessionData, bool) {
	cookies := r.Cookies()
	if len(cookies) == 0 {
		return "", nil, false
	}

	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		log.Printf("GetCookie: %v\n", err)
		return "", nil, false
	}

	s, ok := c.DataMap[cookie.Value]
	if !ok {
		return "", nil, false
	}

	if s.ExpireAt.Before(time.Now()) {
		delete(c.DataMap, cookie.Value)
		err = sqlite.DB.DeleteSession(cookie.Value)
		if err != nil {
			log.Printf("DeleteSession: %v\n", err)
		}
		return "", nil, false
	}

	return cookie.Value, &s, true
}

func (c *Control) Delete(w http.ResponseWriter, id string) {
	delete(c.DataMap, id)

	err := sqlite.DB.DeleteSession(id)
	if err != nil {
		log.Printf("DeleteSession: %v\n", err)
	}

	cookie := http.Cookie{
		Name:   c.cookieName,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
}

func (c *Control) Save(w http.ResponseWriter, id string, sessionData *model.SessionData) {
	expireAt := time.Now().Add(globalconst.TimeToExpire * time.Second)
	cookie := &http.Cookie{
		Path:     "/",
		Name:     c.cookieName,
		Value:    id,
		Expires:  expireAt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	sessionData.ExpireAt = expireAt
	c.DataMap[id] = *sessionData

	err := sqlite.DB.SaveSession(id, sessionData)
	if err != nil {
		log.Printf("SaveSession: %v\n", err)
	}

	http.SetCookie(w, cookie)
}

func (c *Control) Create() (string, *model.SessionData) {
	sessionData := &model.SessionData{
		ExpireAt: time.Now().Add(globalconst.TimeToExpire * time.Second),
	}

	rand := util.RandomID()

	c.DataMap[rand] = *sessionData

	err := sqlite.DB.SaveSession(rand, sessionData)
	if err != nil {
		log.Printf("SaveSession: %v\n", err)
	}

	return rand, sessionData
}

func (c *Control) RemoveExpired() {
	for k, v := range c.DataMap {
		if v.ExpireAt.Before(time.Now()) {
			delete(c.DataMap, k)
		}
	}

	err := sqlite.DB.DeleteExpiredSessions()
	if err != nil {
		log.Printf("DeleteExpiredSessions: %v\n", err)
	}
}
