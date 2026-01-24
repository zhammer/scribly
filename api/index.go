package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"scribly/cmd/site"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

var router http.Handler

func url() string {
	url := os.Getenv("VERCEL_PROJECT_PRODUCTION_URL")
	url, _ = strings.CutPrefix(url, "www.")
	return "https://" + url
}

func init() {
	os.Setenv("SITE_URL", url())

	cfg := site.Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	router, err = site.MakeRouter(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	router.ServeHTTP(w, r)
}
