package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"path"
	"scribly/internal"
	"strconv"
	"text/template"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port             int    `default:"8000"`
	DatabaseURL      string `envconfig:"database_url" default:"postgres://scribly:pass@localhost/scribly?sslmode=disable"`
	SessionSecretKey string `envconfig:"session_secret_key" default:"dev_session_secret"`
}

func main() {
	cfg := Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	router, err := makeRouter(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := http.ListenAndServe(net.JoinHostPort("", strconv.Itoa(cfg.Port)), router); err != nil {
		log.Fatal(err)
	}
}

func makeRouter(cfg Config) (http.Handler, error) {
	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)

	sessions := NewSessionHelper(cfg.SessionSecretKey)

	router := mux.NewRouter()

	opt, err := pg.ParseURL(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(opt)
	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	scribly, err := internal.NewScribly(db, nil)
	if err != nil {
		return nil, err
	}

	indexTmpl := tmpl("index.tmpl")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := indexTmpl.ExecuteTemplate(w, "index.tmpl", nil); err != nil {
			http.Error(w, "Internal Server Error", 500)
		}
	}).Methods("GET")

	loginTmpl := tmpl("login.tmpl")
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := loginTmpl.ExecuteTemplate(w, "login.tmpl", nil); err != nil {
			http.Error(w, "Internal Server Error", 500)
		}
	}).Methods("GET")

	signupTmpl := tmpl("signup.tmpl")
	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if err := signupTmpl.ExecuteTemplate(w, "signup.tmpl", nil); err != nil {
			http.Error(w, "Internal Server Error", 500)
		}
	}).Methods("GET")

	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var input internal.SignUpInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		user, err := scribly.SignUp(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := sessions.SaveUser(user, r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("POST")

	meTmpl := tmpl("me.tmpl")
	router.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		me, err := scribly.Me(r.Context(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := meTmpl.ExecuteTemplate(w, "me.tmpl", ViewData{me, r}); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// refresh the session user in case we get updated user info from db
		_ = sessions.SaveUser(me.User, r, w)
	})

	return router, nil
}

func tmpl(tmplPath string) *template.Template {
	return template.Must(
		template.ParseFiles(
			"gotemplates/_layout.tmpl",
			path.Join("gotemplates", tmplPath),
		),
	)
}
