package main

import (
	"context"
	"fmt"
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
	Debug            bool
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

	if cfg.Debug {
		db.AddQueryHook(DBLogger{})
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

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var input internal.LoginInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		user, err := scribly.LogIn(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := sessions.SaveUser(user, r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)

	}).Methods("POST")

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

		http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)

	}).Methods("POST")

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sessions.ClearUser(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}).Methods("POST", "GET")

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

	storyTmpl := tmpl("story.tmpl")
	addPeopleToStoryTmpl := tmpl("addpeopletostory.tmpl")
	router.HandleFunc("/stories/{storyId:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		story, err := scribly.UserStory(r.Context(), user.ID, storyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch story.Story.State {
		case internal.StoryStateDraft:
			if err := addPeopleToStoryTmpl.ExecuteTemplate(w, "addpeopletostory.tmpl", ViewData{story, r}); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		default:
			if err := storyTmpl.ExecuteTemplate(w, "story.tmpl", ViewData{story, r}); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}

	}).Methods("GET")

	router.HandleFunc("/stories/{storyId:[0-9]+}/addcowriters", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var input internal.AddCowritersInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		err := scribly.AddCowriters(r.Context(), *user, storyID, input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/stories/%d", storyID), http.StatusSeeOther)
	}).Methods("POST")

	router.HandleFunc("/stories/{storyId:[0-9]+}/turn", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var input internal.TurnInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		err := scribly.TakeTurn(r.Context(), *user, storyID, input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/stories/%d", storyID), http.StatusSeeOther)
	}).Methods("POST")

	newStoryTmpl := tmpl("newstory.tmpl")
	router.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if err := newStoryTmpl.ExecuteTemplate(w, "newstory.tmpl", nil); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var input internal.StartStoryInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		story, err := scribly.StartStory(r.Context(), *user, input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/stories/%d", story.ID), http.StatusSeeOther)
	}).Methods("POST")

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
