package main

import (
	"net/http"
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

func (h RedirectHandler) buildRedirectURL(req http.Request) (string, error) {
	target := req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Printf("redirect from: http://%s to: https://%s", target, target)

	return "https://" + target, nil
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.buildRedirectURL(*r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}