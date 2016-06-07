package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bmizerany/assert"
)

func testHerokuProvider(hostname string) *HerokuProvider {
	p := NewHerokuProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
	}
	return p
}

func testHerokuBackend(payload string) *httptest.Server {
	path := "/account"

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			if url.Path != path {
				w.WriteHeader(404)
			} else if r.Header.Get("Authorization") != "Bearer imaginary_access_token" {
				w.WriteHeader(403)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}

func TestHerokuProviderDefaults(t *testing.T) {
	p := testHerokuProvider("")
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Heroku", p.Data().ProviderName)
	assert.Equal(t, "https://id.heroku.com/oauth/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://id.heroku.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://api.heroku.com/account",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://id.heroku.com/oauth/authorizations",
		p.Data().ValidateURL.String())
	assert.Equal(t, "identity", p.Data().Scope)
}

func TestHerokuProviderOverrides(t *testing.T) {
	p := NewHerokuProvider(
		&ProviderData{
			LoginURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/auth"},
			RedeemURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/token"},
			ProfileURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/profile"},
			ValidateURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/tokeninfo"},
			Scope: "profile"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Heroku", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/auth",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/oauth/profile",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://example.com/oauth/tokeninfo",
		p.Data().ValidateURL.String())
	assert.Equal(t, "profile", p.Data().Scope)
}

func TestHerokuProviderGetEmailAddress(t *testing.T) {
	b := testHerokuBackend(`{"email": "user@heroku.com"}`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testHerokuProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user@heroku.com", email)
}

func TestHerokuProviderGetEmailAddressFailedRequest(t *testing.T) {
	b := testHerokuBackend("unused payload")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testHerokuProvider(b_url.Host)

	// We'll trigger a request failure by using an unexpected access
	// token. Alternatively, we could allow the parsing of the payload as
	// JSON to fail.
	session := &SessionState{AccessToken: "unexpected_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}
