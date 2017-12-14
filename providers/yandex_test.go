package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testYandexProvider(hostname string) *YandexProvider {
	p := NewYandexProvider(
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
		updateURL(p.Data().ValidateURL, hostname)
	}
	return p
}

func testYandexBackend(payload string) *httptest.Server {
	path := "/info"
	query := "oauth_token=imaginary_access_token"

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			uri := r.URL
			if uri.Path != path || uri.RawQuery != query {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}

func TestYandexProviderDefaults(t *testing.T) {
	p := testYandexProvider("")
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Yandex", p.Data().ProviderName)
	assert.Equal(t, "https://oauth.yandex.com/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://oauth.yandex.com/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://login.yandex.ru/info",
		p.Data().ProfileURL.String())
	assert.Equal(t, "login:email", p.Data().Scope)
}

func TestYandexProviderOverrides(t *testing.T) {
	p := NewYandexProvider(
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
				Path:   "/info"},
			Scope: "profile"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Yandex", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/auth",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/info",
		p.Data().ProfileURL.String())
	assert.Equal(t, "profile", p.Data().Scope)
}

func TestYandexProviderGetEmailAddress(t *testing.T) {
	b := testYandexBackend("{\"default_email\": \"michael.bland@gsa.gov\"}")
	defer b.Close()

	bUrl, _ := url.Parse(b.URL)
	p := testYandexProvider(bUrl.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "michael.bland@gsa.gov", email)
}

// Note that trying to trigger the "failed building request" case is not
// practical, since the only way it can fail is if the URL fails to parse.
func TestYandexProviderGetEmailAddressFailedRequest(t *testing.T) {
	b := testYandexBackend("unused payload")
	defer b.Close()

	bUrl, _ := url.Parse(b.URL)
	p := testYandexProvider(bUrl.Host)

	// We'll trigger a request failure by using an unexpected access
	// token. Alternatively, we could allow the parsing of the payload as
	// JSON to fail.
	session := &SessionState{AccessToken: "unexpected_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}

func TestYandexProviderGetEmailAddressEmailNotPresentInPayload(t *testing.T) {
	b := testYandexBackend("{\"foo\": \"bar\"}")
	defer b.Close()

	bUrl, _ := url.Parse(b.URL)
	p := testYandexProvider(bUrl.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}
