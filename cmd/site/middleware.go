package main

import (
	"net/http"
	"net/url"
)

func HerokuHTTPSMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-FORWARDED-PROTO") == "http" {
			urlCopy, _ := url.Parse(r.URL.String())
			urlCopy.Scheme = "https"
			http.Redirect(w, r, urlCopy.String(), http.StatusTemporaryRedirect)
			return
		}
		h.ServeHTTP(w, r)
	})
}
