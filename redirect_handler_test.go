package main

import (
	"net/url"
	"testing"

	"github.com/bmizerany/assert"
)

func testRedirectorOptions() *Options {
	o := NewOptions()
	o.Upstreams = append(o.Upstreams, "http://127.0.0.1:8080/")
	o.CookieSecret = "foobar"
	o.ClientID = "bazquux"
	o.ClientSecret = "xyzzyplugh"
	o.EmailDomains = []string{"*"}
	o.RedirectURL = "https://external.com/oauth2/callback"
	return o
}

func TestBuildRedirectURL(t *testing.T) {
	opts := testRedirectorOptions()
	url, _ := url.Parse("http://internal.com:9000/barkis?a=1&b=2")
	h := NewRedirectHandler(*opts)
	out, _ := h.buildRedirectURL(*url)

	assert.Equal(t, out, "https://external.com/barkis?a=1&b=2")
}

func TestBuildRedirectURLUserAndPass(t *testing.T) {
	opts := testRedirectorOptions()
	url, _ := url.Parse("http://charles:dickens@internal.com:9000/boffin?a=1&b=2")
	h := NewRedirectHandler(*opts)
	out, _ := h.buildRedirectURL(*url)

	assert.Equal(t, out, "https://charles:dickens@external.com/boffin?a=1&b=2")
}

func TestBuildRedirectURLFragment(t *testing.T) {
	opts := testRedirectorOptions()
	url, _ := url.Parse("http://internal.com:9000/boffin/#test")
	h := NewRedirectHandler(*opts)
	out, _ := h.buildRedirectURL(*url)

	assert.Equal(t, out, "https://external.com/boffin/#test")
}
