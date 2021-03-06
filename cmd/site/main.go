package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	embedded "scribly"
	"scribly/internal"
	"strconv"

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

	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(embedded.StaticFS)))

	router.HandleFunc("/_cypress_email", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/tempfile.html")
	})

	errorPage := func(w http.ResponseWriter, r *http.Request, e error) error {
		fmt.Printf("error on request %s: %s\n", r.URL.String(), e.Error())
		return embedded.WebTemplates.ExecuteTemplate(w, "exception.tmpl", ViewData{e, r, "Uh Oh!", ""})
	}

	router.HandleFunc("/exception", func(w http.ResponseWriter, r *http.Request) {
		err := fmt.Errorf("Raising an exception, intentionally!")
		errorPage(w, r, err)
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := embedded.WebTemplates.ExecuteTemplate(w, "index.tmpl", ViewData{nil, r, "", ""}); err != nil {
			errorPage(w, r, err)
		}
	}).Methods("GET")

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if user, _ := sessions.GetUser(r); user != nil {
			http.Redirect(w, r, "/me", http.StatusTemporaryRedirect)
			return
		}

		if err := embedded.WebTemplates.ExecuteTemplate(w, "login.tmpl", ViewData{nil, r, "Scribly - Log in", ""}); err != nil {
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

	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if err := embedded.WebTemplates.ExecuteTemplate(w, "signup.tmpl", ViewData{nil, r, "Scribly - Sign up", ""}); err != nil {
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

		if err := embedded.WebTemplates.ExecuteTemplate(w, "me.tmpl", ViewData{me, r, "", ""}); err != nil {
			errorPage(w, r, err)
			return
		}

		// refresh the session user in case we get updated user info from db
		_ = sessions.SaveUser(me.User, r, w)
	})

	router.HandleFunc("/stories/{storyId:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		storyID, _ := strconv.Atoi(mux.Vars(r)["storyId"])

		story, err := scribly.UserStory(r.Context(), *user, storyID)
		if err != nil {
			errorPage(w, r, err)
			return
		}

		switch story.Story.State {
		case internal.StoryStateDraft:
			userSuggestions, _ := scribly.UserSuggestions(r.Context(), *user)
			viewData := ViewData{
				Data: map[string]interface{}{
					"UserStory":       story,
					"UserSuggestions": userSuggestions,
				},
				Request: r,
				title:   fmt.Sprintf("%s - Add Cowriters", story.Story.Title),
			}
			if err := embedded.WebTemplates.ExecuteTemplate(w, "addpeopletostory.tmpl", viewData); err != nil {
				errorPage(w, r, err)
				return
			}
		default:
			if err := embedded.WebTemplates.ExecuteTemplate(w, "story.tmpl", ViewData{story, r, fmt.Sprintf("Scribly - %s", story.Story.Title), ""}); err != nil {
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

	router.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		user, _ := sessions.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if err := embedded.WebTemplates.ExecuteTemplate(w, "newstory.tmpl", ViewData{nil, r, "Scribly - New Story", "page page-tall"}); err != nil {
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
			title: "Nudge delivered",
		}
		if err := embedded.WebTemplates.ExecuteTemplate(w, "nudged.tmpl", data); err != nil {
			errorPage(w, r, err)
			return
		}

	})

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

		if err := embedded.WebTemplates.ExecuteTemplate(w, "emailverificationrequested.tmpl", ViewData{user, r, "Email Verification Sent", ""}); err != nil {
			errorPage(w, r, err)
			return
		}
	}).Methods("POST")

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

		if err := embedded.WebTemplates.ExecuteTemplate(w, "emailverificationsuccess.tmpl", ViewData{user, r, "Email Verified!", ""}); err != nil {
			errorPage(w, r, err)
			return
		}
	}).Methods("GET")

	// middleware
	router.Use(HerokuHTTPSMiddleware)

	return router, nil
}

func pathFromURL(urlString string) string {
	url, _ := url.Parse(urlString)
	return url.Path + "?" + url.RawQuery
}
