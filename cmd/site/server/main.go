package main

import (
	"log"
	"net"
	"net/http"
	"scribly/cmd/site"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	cfg := site.Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	router, err := site.MakeRouter(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting server on port %d", cfg.Port)
	if err := http.ListenAndServe(net.JoinHostPort("", strconv.Itoa(cfg.Port)), router); err != nil {
		log.Fatal(err)
	}
}
