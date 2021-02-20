package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
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
	SendgridBaseURL  string `envconfig:"sendgrid_base_url" default:"https://api.sendgrid.com"`
	SendgridAPIKey   string `envconfig:"sendgrid_api_key" default:"test_sendgrid_api_key"`
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

	sendgrid := internal.NewSendgridClient(cfg.SendgridBaseURL, cfg.SendgridAPIKey)
	messageGateway := internal.GoroutineMessageGateway{}

	scribly, err := internal.NewScribly(db, sendgrid, &messageGateway)
	if err != nil {
		return nil, err
	}
	messageGateway.Scribly(scribly)

	staticDir := "/static/"
	router.
		PathPrefix(staticDir).
		Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("."+staticDir))))

	exceptionTmpl := tmpl("exception.tmpl")
	errorPage := func(w http.ResponseWriter, r *http.Request, e error) error {
		fmt.Printf("error on request %s: %s\n", r.URL.String(), e.Error())
		return exceptionTmpl.ExecuteTemplate(w, "exception.tmpl", ViewData{e, r})
	}

	router.HandleFunc("/exception", func(w http.ResponseWriter, r *http.Request) {
		err := fmt.Errorf("Raising an exception, intentionally!")
		errorPage(w, r, err)
	})

	indexTmpl := tmpl("index.tmpl")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := indexTmpl.ExecuteTemplate(w, "index.tmpl", nil); err != nil {
			errorPage(w, r, err)
		}
	}).Methods("GET")

	loginTmpl := tmpl("login.tmpl")
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := loginTmpl.ExecuteTemplate(w, "login.tmpl", nil); err != nil {
			errorPage(w, r, err)
		}
	}).Methods("GET")

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := r.ParseForm(); err != nil {
			errorPage(w, r, err)
			return
		}

		var input internal.LoginInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			errorPage(w, r, err)
			return
		}

		user, err := scribly.LogIn(r.Context(), input)
		if err != nil {
			errorPage(w, r, err)
			return
		}

		if err := sessions.SaveUser(user, r, w); err != nil {
			errorPage(w, r, err)
			return
		}

		http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)

	}).Methods("POST")

	signupTmpl := tmpl("signup.tmpl")
	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if err := signupTmpl.ExecuteTemplate(w, "signup.tmpl", nil); err != nil {
			errorPage(w, r, err)
		}
	}).Methods("GET")

	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := r.ParseForm(); err != nil {
			errorPage(w, r, err)
			return
		}

		var input internal.SignUpInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			errorPage(w, r, err)
			return
		}

		user, err := scribly.SignUp(r.Context(), input)
		if err != nil {
			errorPage(w, r, err)
			return
		}

		if err := sessions.SaveUser(user, r, w); err != nil {
			errorPage(w, r, err)
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
			errorPage(w, r, err)
			return
		}

		if err := meTmpl.ExecuteTemplate(w, "me.tmpl", ViewData{me, r}); err != nil {
			errorPage(w, r, err)
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
			errorPage(w, r, err)
			return
		}

		switch story.Story.State {
		case internal.StoryStateDraft:
			if err := addPeopleToStoryTmpl.ExecuteTemplate(w, "addpeopletostory.tmpl", ViewData{story, r}); err != nil {
				errorPage(w, r, err)
				return
			}
		default:
			if err := storyTmpl.ExecuteTemplate(w, "story.tmpl", ViewData{story, r}); err != nil {
				errorPage(w, r, err)
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
			errorPage(w, r, err)
			return
		}

		var input internal.AddCowritersInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			errorPage(w, r, err)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		err := scribly.AddCowriters(r.Context(), *user, storyID, input)
		if err != nil {
			errorPage(w, r, err)
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
			errorPage(w, r, err)
			return
		}

		var input internal.TurnInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			errorPage(w, r, err)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		err := scribly.TakeTurn(r.Context(), *user, storyID, input)
		if err != nil {
			errorPage(w, r, err)
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
			errorPage(w, r, err)
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
			errorPage(w, r, err)
			return
		}

		var input internal.StartStoryInput
		if err := formDecoder.Decode(&input, r.PostForm); err != nil {
			errorPage(w, r, err)
			return
		}

		story, err := scribly.StartStory(r.Context(), *user, input)
		if err != nil {
			errorPage(w, r, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/stories/%d", story.ID), http.StatusSeeOther)
	}).Methods("POST")

	router.HandleFunc("/stories/{storyId:[0-9]+}/{action:hide|unhide}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])
		action := mux.Vars(r)["action"]

		hiddenStatus := internal.HiddenStatusHidden
		if action == "unhide" {
			hiddenStatus = internal.HiddenStatusUnhidden
		}

		if err := scribly.Hide(r.Context(), *user, storyID, hiddenStatus); err != nil {
			errorPage(w, r, err)
			return
		}

		http.Redirect(w, r, pathFromURL(r.Header.Get("referer")), http.StatusSeeOther)

	}).Methods("POST")

	nudgedTmpl := tmpl("nudged.tmpl")
	router.HandleFunc("/stories/{storyId:[0-9]+}/nudge/{nudgeeId:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])
		nudgeeID, _ := strconv.Atoi(mux.Vars(r)["nudgeeId"])

		if err := scribly.Nudge(r.Context(), *user, nudgeeID, storyID); err != nil {
			errorPage(w, r, err)
			return
		}

		data := ViewData{
			Request: r,
			Data: map[string]interface{}{
				"User":    user,
				"StoryID": storyID,
			},
		}
		if err := nudgedTmpl.ExecuteTemplate(w, "nudged.tmpl", data); err != nil {
			errorPage(w, r, err)
			return
		}

	})

	emailVerificationRequestedTmpl := tmpl("emailverificationrequested.tmpl")
	router.HandleFunc("/email-verification", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if err := scribly.SendVerificationEmail(r.Context(), user.ID); err != nil {
			errorPage(w, r, err)
			return
		}

		if err := emailVerificationRequestedTmpl.ExecuteTemplate(w, "emailverificationrequested.tmpl", ViewData{user, r}); err != nil {
			errorPage(w, r, err)
			return
		}
	}).Methods("POST")

	emailVerificationSuccessTmpl := tmpl("emailverificationsuccess.tmpl")
	router.HandleFunc("/email-verification", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		token := r.URL.Query().Get("token")

		if err := scribly.VerifyEmail(r.Context(), *user, token); err != nil {
			errorPage(w, r, err)
			return
		}

		if err := emailVerificationSuccessTmpl.ExecuteTemplate(w, "emailverificationsuccess.tmpl", ViewData{user, r}); err != nil {
			errorPage(w, r, err)
			return
		}
	}).Methods("GET")

	return router, nil
}

func pathFromURL(urlString string) string {
	url, _ := url.Parse(urlString)
	return url.Path + "?" + url.RawQuery
}

func tmpl(tmplPath string) *template.Template {
	return template.Must(
		template.ParseFiles(
			"templates/_layout.tmpl",
			path.Join("templates", tmplPath),
		),
	)
}
