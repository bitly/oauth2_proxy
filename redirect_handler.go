package main

import (
	"net/http"
	"net/url"
	"log"
)

type RedirectHandler struct {
	Opts Options
}

func NewRedirectHandler(opts Options) RedirectHandler {
	return RedirectHandler{
		Opts: opts,
	}
}

func (h RedirectHandler) buildRedirectURL(requestURL url.URL) (string, error) {
	target := "https://" + requestURL.Host + requestURL.Path
	if len(requestURL.RawQuery) > 0 {
		target += "?" + requestURL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	return target, nil
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.buildRedirectURL(*r.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}