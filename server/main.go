package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"realm/globalconst"
	"realm/handler"
	"realm/model"
	"realm/session"
	"realm/sqlite"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
)

type Config struct {
	GithubClientID     string `ini:"github_client_id" cfg:"github_client_id" cfgRequired:"true" cfgHelper:"Github Client ID"`
	GithubClientSecret string `ini:"github_client_secret" cfg:"github_client_secret" cfgRequired:"true" cfgHelper:"Github Client Secret"`
	GithubCallbackURL  string `ini:"github_callback_url" cfg:"github_callback_url" cfgRequired:"true" cfgHelper:"Github Callback URL"`
	DatabaseName       string `ini:"database_name" cfg:"database_name" cfgRequired:"true" cfgHelper:"Database Name"`
	Port               int    `ini:"port" cfg:"port" cfgDefault:"8080" cfgHelper:"Port"`
}

var (
	//go:embed assets/*
	assets embed.FS
)

// ///////////////////////////////////
func forumHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homeHandler")

	log.Println("r.URL.Path:", r.URL.Path)

	sid, sd, ok := session.SC.Get(r)
	if !ok {
		log.Println("1 session not found")
		sid, sd = session.SC.Create()
	}

	if !sd.LoggedIn {
		log.Println("not logged in")
	}
	log.Println("name:", sd.UserName)

	// renew session
	session.SC.Save(w, sid, sd)

	log.Println("sid:", sid)
	//////////////////////////

	index, err := assets.ReadFile("assets/forum.html")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("forum.html").Parse(string(index))
	if err != nil {
		log.Fatal(err)
	}

	fl, err := sqlite.DB.GetForumList()
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		SessionData    *model.SessionData
		GitHubLoginURL string
		LogoutURL      string
		ForumList      []model.Forum
	}{
		SessionData:    sd,
		GitHubLoginURL: "/forum/github/login",
		LogoutURL:      "/forum/logout",
		ForumList:      fl,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("logoutHandler")
	sid, sd, ok := session.SC.Get(r)
	if !ok {
		http.Redirect(w, r, "/forum", http.StatusFound)
		return
	}

	sd.LoggedIn = false

	// remove session
	session.SC.Delete(w, sid)

	http.Redirect(w, r, "/forum", http.StatusFound)
}

func issueSession() http.Handler {
	log.Println("issueSession")
	fn := func(w http.ResponseWriter, r *http.Request) {
		sid, sd, ok := session.SC.Get(r)
		if !ok {
			log.Println("2 session not found")
			sid, sd = session.SC.Create()
		}

		ctx := r.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: validar todos os campos, se for nil trocar por um valor padrão
		log.Println("sid:", sid)
		//log.Printf("githubUser: %#v\n", githubUser)
		log.Println("ID:", *githubUser.ID)
		log.Println("Login:", *githubUser.Login)

		/////////////////
		// save user data
		user := model.User{
			ID:            sd.UserID,
			UserName:      *githubUser.Name,
			AvatarURL:     *githubUser.AvatarURL,
			OAuthProvider: "github",
			OAuthUserID:   fmt.Sprintf("%d", *githubUser.ID),
		}

		user, err = sqlite.DB.SaveUser(&user)
		if err != nil {
			log.Println(err)
		}

		///////////////////
		// save session data
		sdAUX := model.SessionData{
			OAuthProvider: "github",
			OAuthUserID:   fmt.Sprintf("%d", *githubUser.ID),
			UserName:      *githubUser.Name,
			AvatarURL:     *githubUser.AvatarURL,
			SessionID:     sid,
			LoggedIn:      true,
			ExpireAt:      time.Now().Add(globalconst.TimeToExpire * time.Second),
			UserID:        user.ID,
		}
		sd = &sdAUX

		log.Println("name:", sdAUX.UserName)
		// renew session
		session.SC.Save(w, sid, sd)

		http.Redirect(w, r, "/forum", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

/////////////////////////////////////

func main() {
	cfg := Config{}

	config.File = "config.ini"
	err := config.Parse(&cfg)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("database name: %s\n", cfg.DatabaseName)
	err = sqlite.Open(cfg.DatabaseName)
	if err != nil {
		log.Fatal(err)
	}

	err = sqlite.DB.CreateSessionTables()
	if err != nil {
		log.Fatal(err)
	}

	err = sqlite.DB.CreateForumTables()
	if err != nil {
		log.Fatal(err)
	}

	err = sqlite.DB.CreateChatTables()
	if err != nil {
		log.Fatal(err)
	}

	err = sqlite.DB.CreateUserTables()
	if err != nil {
		log.Fatal(err)
	}

	session.New(globalconst.CookieName)

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			session.SC.RemoveExpired()
		}
	}()

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubCallbackURL,
		Endpoint:     githubOAuth2.Endpoint,
	}

	// state param cookies require HTTPS by default; disable for localhost development
	//stateConfig := gologin.DebugOnlyCookieConfig
	stateConfig := gologin.DefaultCookieConfig

	assetsRFS, _ := fs.Sub(assets, "assets")
	var assetsFS = http.FS(assetsRFS)

	fs := http.FileServer(assetsFS)

	mux := http.NewServeMux()

	mux.HandleFunc("/realm/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/ws", handler.Websocket)
	mux.HandleFunc("/forum/", forumHandler)
	mux.HandleFunc("/forum/logout", logoutHandler)

	mux.Handle(
		"/forum/github/login",
		github.StateHandler(
			stateConfig,
			github.LoginHandler(oauth2Config, nil)))
	mux.Handle(
		"/forum/github/callback",
		github.StateHandler(
			stateConfig,
			github.CallbackHandler(oauth2Config, issueSession(), nil)))

	// recebe post de usuário
	mux.HandleFunc("/forum/post", func(w http.ResponseWriter, r *http.Request) {
		log.Println("postHandler")
		sid, sd, ok := session.SC.Get(r)
		if !ok {
			http.Redirect(w, r, "/forum", http.StatusFound)
			return
		}

		if sd.UserName != "" {
			log.Println("name:", sd.UserName)
		}

		post := ""
		if r.Method == http.MethodPost {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				return
			}
			post = string(b)
		}

		log.Println("sid:", sid)
		log.Println("post:", post)

	})

	s := &http.Server{
		Handler:        mux,
		Addr:           fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on port %d\n", cfg.Port)
	log.Fatal(s.ListenAndServe())
}
