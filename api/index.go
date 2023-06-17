package handler

import (
	"fmt"
	"log"
	"net/http"
	"scribly/cmd/site"

	"github.com/kelseyhightower/envconfig"
)

var router http.Handler

func init() {
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
