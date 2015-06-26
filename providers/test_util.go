package providers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
)

func NewTestQueryBackend(path, query, payload string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			if url.Path != path || url.RawQuery != query {
				log.Printf("unexpected request:\n"+
					"  expected:  %s?%s\n"+
					"  actual:    %s?%s", path, query,
					url.Path, url.RawQuery)
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}

func NewTestPostBackend(path string, form url.Values, payload string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			r.ParseForm()
			if url.Path != path || reflect.DeepEqual(r.Form, form) {
				log.Printf("unexpected request:\n"+
					"  expected:  %s\n    %v\n"+
					"  actual:    %s\n    %v", path, form,
					url.Path, r.Form)
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}
