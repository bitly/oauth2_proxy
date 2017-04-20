package providers

import (
	"github.com/bmizerany/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testZendeskProvider(hostname string) *ZendeskProvider {
	p := NewZendeskProvider(
		&ProviderData{
			ProviderName:      "",
			LoginURL:          &url.URL{},
			RedeemURL:         &url.URL{},
			ProfileURL:        &url.URL{},
			ValidateURL:       &url.URL{},
			Scope:             ""})
	p.Configure("example")
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}
	return p
}

func TestZendeskProviderOverrides(t *testing.T) {
	p := NewZendeskProvider(
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
			ProtectedResource: &url.URL{
				Scheme: "https",
				Host:   "example.com"},
			Scope: "profile"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Zendesk", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/auth",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/oauth/profile",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://example.com/oauth/tokeninfo",
		p.Data().ValidateURL.String())
	assert.Equal(t, "https://example.com",
		p.Data().ProtectedResource.String())
	assert.Equal(t, "profile", p.Data().Scope)
}

func TestZendeskSetSubdomain(t *testing.T) {
	p := testZendeskProvider("")
	p.Configure("example")
	assert.Equal(t, "Zendesk", p.Data().ProviderName)
	assert.Equal(t, "example", p.Subdomain)
	assert.Equal(t, "https://example.zendesk.com/oauth/authorizations/new",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.zendesk.com/oauth/tokens",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.zendesk.com/api/v2/users/me.json",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://example.zendesk.com",
		p.Data().ProtectedResource.String())
	assert.Equal(t, "",
		p.Data().ValidateURL.String())
	assert.Equal(t, "read", p.Data().Scope)
}

func testZendeskBackend(payload string) *httptest.Server {
	path := "/api/v2/users/me.json"

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

func TestZendeskProviderGetEmailAddress(t *testing.T) {
	b := testZendeskBackend(`{"user": {"id":5383137307,"url":"https://example.zendesk.com/api/v2/end_users/5555555555.json","name":"Zensdesk Test","email":"zendesk.test@example.com","created_at":"2016-03-31T14:51:17Z","updated_at":"2016-04-04T09:05:00Z","time_zone":"Eastern Time (US & Canada)","phone":null,"photo":null,"locale_id":1,"locale":"en-US","organization_id":null,"role":"end-user","verified":true}}`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testZendeskProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "zendesk.test@example.com", email)
}
