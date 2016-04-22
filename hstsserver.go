package main

import (
	"net/http"
	"strings"
)

type HSTSServer struct{}

func (server HSTSServer) Handle(pattern string, handler http.Handler) {
	server.HandleFunc(pattern, handler.ServeHTTP)
}

func (HSTSServer) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if doHSTS(w, r) {
			handlerFunc(w, r)
		}
	})
}

// `cont` indicates whether the callee should continue further processing.
func doHSTS(w http.ResponseWriter, r *http.Request) (cont bool) {
	if strings.HasPrefix(r.Host, "appspot.com") {
		if r.TLS == nil {
			http.Redirect(w, r, "https://"+r.Host+r.URL.Path, 301)
			return false
		} else {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
	}
	return true
}
