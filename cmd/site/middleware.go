package main

import (
	"net/http"
)

func HerokuHTTPSMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-FORWARDED-PROTO") == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusTemporaryRedirect)
			return
		}
		h.ServeHTTP(w, r)
	})
}
