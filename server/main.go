package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"realm/db/sqlite"
	"realm/handler"
	"realm/session"

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

type sessionData struct {
	OAuthProvider string
	UserName      string
	AvatarURL     string
	SessionID     string
}

var (
	sc *session.Control

	//go:embed assets/*
	assets embed.FS
)

// ///////////////////////////////////
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("homeHandler")

	sid, sd, ok := sc.Get(r)
	if !ok {
		fmt.Println("1 session not found")
		sid, sd = sc.Create()
	}

	if sd.Data != nil {
		sdAUX, ok := sd.Data.(*sessionData)
		if !ok {
			log.Fatal("type assertion failed sessionData")
		}
		fmt.Println("name:", sdAUX.UserName)
	} else {
		fmt.Println("sd.Data is nil")
	}

	// renew session
	sc.Save(w, sid, sd)

	fmt.Println("sid:", sid)
	//////////////////////////

	index, err := assets.ReadFile("assets/forum.html")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("forum.html").Parse(string(index))
	if err != nil {
		log.Fatal(err)
	}

	var (
		sdAUX *sessionData
	)

	if sd.Data != nil {
		sdAUX, ok = sd.Data.(*sessionData)
		if !ok {
			log.Fatal("type assertion failed sessionData")
		}
	}
	data := struct {
		SessionData    *sessionData
		GitHubLoginURL string
		LogoutURL      string
	}{
		SessionData:    sdAUX,
		GitHubLoginURL: "/forum/github/login",
		LogoutURL:      "/forum/logout",
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logoutHandler")
	sid, sd, ok := sc.Get(r)
	if !ok {
		http.Redirect(w, r, "/forum", http.StatusFound)
		return
	}

	sd.Data = nil

	// remove session
	sc.Delete(w, sid)

	http.Redirect(w, r, "/forum", http.StatusFound)
}

func issueSession() http.Handler {
	fmt.Println("issueSession")
	fn := func(w http.ResponseWriter, r *http.Request) {
		sid, sd, ok := sc.Get(r)
		if !ok {
			fmt.Println("2 session not found")
			sid, sd = sc.Create()
		}

		ctx := r.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("sid:", sid)
		fmt.Println("githubUser id.........:", *githubUser.ID)
		fmt.Println("githubUser login......:", *githubUser.Login)
		fmt.Println("githubUser email......:", *githubUser.Email)
		fmt.Println("githubUser name.......:", *githubUser.Name)
		fmt.Println("githubUser avatar.....:", *githubUser.AvatarURL)
		fmt.Println("githubUser url........:", *githubUser.URL)
		fmt.Println("githubUser html url...:", *githubUser.HTMLURL)
		fmt.Println("githubUser followers..:", *githubUser.Followers)
		fmt.Println("githubUser following..:", *githubUser.Following)
		fmt.Println("githubUser created at.:", *githubUser.CreatedAt)

		sdAUX := sessionData{
			OAuthProvider: "github",
			UserName:      *githubUser.Name,
			AvatarURL:     *githubUser.AvatarURL,
			SessionID:     sid,
		}
		sd.Data = &sdAUX

		fmt.Println("name:", sdAUX.UserName)
		// renew session
		sc.Save(w, sid, sd)

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
		fmt.Println(err)
		return
	}

	log.Printf("database name: %s\n", cfg.DatabaseName)
	err = sqlite.New(cfg.DatabaseName)
	if err != nil {
		log.Fatal(err)
	}

	const cookieName = "forum_session"
	sc = session.New(cookieName)

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sc.RemoveExpired()
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
	//mux.HandleFunc("/forum/", handler.Forum)

	mux.HandleFunc("/forum/", homeHandler)
	//mux.HandleFunc("/login", loginHandler)
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
		fmt.Println("postHandler")
		sid, sd, ok := sc.Get(r)
		if !ok {
			http.Redirect(w, r, "/forum", http.StatusFound)
			return
		}

		sdAUX, ok := sd.Data.(*sessionData)
		if !ok {
			log.Fatal("type assertion failed sessionData")
		}
		fmt.Println("name:", sdAUX.UserName)

		var post string
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("sid:", sid)
		fmt.Println("post:", post)
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
