package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"scribly/cmd/site"

	"github.com/kelseyhightower/envconfig"
)

var router http.Handler

func init() {
	// https://github.com/vercel-community/php/issues/603#issuecomment-3792299533
	os.Setenv("AWS_LAMBDA_EXEC_WRAPPER_OLD", os.Getenv("AWS_LAMBDA_EXEC_WRAPPER"))
	os.Setenv("AWS_LAMBDA_EXEC_WRAPPER", "")
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
	if r.URL.Path == "/env" {
		fmt.Println("Wrapper: ", os.Getenv("AWS_LAMBDA_EXEC_WRAPPER_OLD"))
	}
	router.ServeHTTP(w, r)
}
