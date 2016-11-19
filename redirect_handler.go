package main

import (
	"net/http"
	"net/url"
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
	u, err := url.Parse(h.Opts.RedirectURL)
	if err != nil {
		return "", err
	}
	requestURL.Scheme = u.Scheme
	requestURL.Host = u.Host

	return requestURL.String(), nil
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.buildRedirectURL(*r.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}
