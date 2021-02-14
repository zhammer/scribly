package main

import (
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"scribly/internal"
	"strconv"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port        int    `default:"8000"`
	DatabaseURL string `envconfig:"database_url" default:"postgres://scribly:pass@localhost/scribly?sslmode=disable"`
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
	router := mux.NewRouter()
	templates, err := template.ParseGlob("gotemplates/**")
	if err != nil {
		return nil, err
	}

	opt, err := pg.ParseURL(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(opt)
	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	_, err = internal.NewScribly(db, nil)
	if err != nil {
		return nil, err
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "index.tmpl", nil); err != nil {
			http.Error(w, "Internal Server Error", 500)
		}
	}).Methods("GET")

	return router, nil
}
