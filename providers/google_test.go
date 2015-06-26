package providers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/bmizerany/assert"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newGoogleProvider() *GoogleProvider {
	return NewGoogleProvider(
		&ProviderData{
			ProviderName: "",
			LoginUrl:     &url.URL{},
			RedeemUrl:    &url.URL{},
			ProfileUrl:   &url.URL{},
			ValidateUrl:  &url.URL{},
			Scope:        ""})
}

func testGoogleRedeemBackend(data *ProviderData, redirect_uri string, code string, payload string) (server *httptest.Server) {
	path := "/oauth2/v3/token"
	form := url.Values{
		"redirect_uri":  {redirect_uri},
		"client_id":     {"0"},
		"client_secret": {""},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}
	server = NewTestPostBackend(path, form, payload)
	data.RedeemUrl, _ = url.Parse(server.URL)
	data.RedeemUrl.Path = path
	return
}

func testGoogleRedeemRefreshTokenBackend(data *ProviderData, refresh_token string, payload string) (server *httptest.Server) {
	path := "/oauth2/v3/token"
	form := url.Values{
		"client_id":     {"0"},
		"client_secret": {""},
		"refresh_token": {refresh_token},
		"grant_type":    {"refresh_token"},
	}
	server = NewTestPostBackend(path, form, payload)
	data.RedeemUrl, _ = url.Parse(server.URL)
	data.RedeemUrl.Path = path
	return
}

func testGoogleValidateTokenBackend(data *ProviderData, access_token string, payload string) (server *httptest.Server) {
	path := "/oauth2/v1/tokeninfo"
	query := "access_token=" + access_token
	server = NewTestQueryBackend(path, query, payload)
	data.ValidateUrl, _ = url.Parse(server.URL)
	data.ValidateUrl.Path = path
	return
}

func TestGoogleProviderDefaults(t *testing.T) {
	p := newGoogleProvider()
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Google", p.Data().ProviderName)
	assert.Equal(t, "https://accounts.google.com/o/oauth2/auth?access_type=offline",
		p.Data().LoginUrl.String())
	assert.Equal(t, "https://www.googleapis.com/oauth2/v3/token",
		p.Data().RedeemUrl.String())
	assert.Equal(t, "https://www.googleapis.com/oauth2/v1/tokeninfo",
		p.Data().ValidateUrl.String())
	assert.Equal(t, "", p.Data().ProfileUrl.String())
	assert.Equal(t, "profile email", p.Data().Scope)
}

func TestGoogleProviderOverrides(t *testing.T) {
	p := NewGoogleProvider(
		&ProviderData{
			LoginUrl: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/auth"},
			RedeemUrl: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/token"},
			ProfileUrl: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/profile"},
			ValidateUrl: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/tokeninfo"},
			Scope: "profile"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Google", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/auth",
		p.Data().LoginUrl.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemUrl.String())
	assert.Equal(t, "https://example.com/oauth/profile",
		p.Data().ProfileUrl.String())
	assert.Equal(t, "https://example.com/oauth/tokeninfo",
		p.Data().ValidateUrl.String())
	assert.Equal(t, "profile", p.Data().Scope)
}

func TestGoogleProviderGetEmailAddress(t *testing.T) {
	p := newGoogleProvider()
	body, err := json.Marshal(
		struct {
			IdToken string `json:"id_token"`
		}{
			IdToken: "ignored prefix." + base64.URLEncoding.EncodeToString([]byte(`{"email": "michael.bland@gsa.gov"}`)),
		},
	)
	assert.Equal(t, nil, err)
	email, err := p.GetEmailAddress(body, "ignored access_token")
	assert.Equal(t, "michael.bland@gsa.gov", email)
	assert.Equal(t, nil, err)
}

func TestGoogleProviderGetEmailAddressInvalidEncoding(t *testing.T) {
	p := newGoogleProvider()
	body, err := json.Marshal(
		struct {
			IdToken string `json:"id_token"`
		}{
			IdToken: "ignored prefix." + `{"email": "michael.bland@gsa.gov"}`,
		},
	)
	assert.Equal(t, nil, err)
	email, err := p.GetEmailAddress(body, "ignored access_token")
	assert.Equal(t, "", email)
	assert.NotEqual(t, nil, err)
}

func TestGoogleProviderGetEmailAddressInvalidJson(t *testing.T) {
	p := newGoogleProvider()

	body, err := json.Marshal(
		struct {
			IdToken string `json:"id_token"`
		}{
			IdToken: "ignored prefix." + base64.URLEncoding.EncodeToString([]byte(`{"email": michael.bland@gsa.gov}`)),
		},
	)
	assert.Equal(t, nil, err)
	email, err := p.GetEmailAddress(body, "ignored access_token")
	assert.Equal(t, "", email)
	assert.NotEqual(t, nil, err)
}

func TestGoogleProviderGetEmailAddressEmailMissing(t *testing.T) {
	p := newGoogleProvider()
	body, err := json.Marshal(
		struct {
			IdToken string `json:"id_token"`
		}{
			IdToken: "ignored prefix." + base64.URLEncoding.EncodeToString([]byte(`{"not_email": "missing"}`)),
		},
	)
	assert.Equal(t, nil, err)
	email, err := p.GetEmailAddress(body, "ignored access_token")
	assert.Equal(t, "", email)
	assert.NotEqual(t, nil, err)
}

func TestGoogleProviderReedeem(t *testing.T) {
	p, redirect_uri, code := newGoogleProvider(), "/redirect", "my code"
	payload := `{"access_token":"access_token",` +
		` "expires_in":3920,` +
		` "token_type":"Bearer",` +
		` "refresh_token":"refresh_token"}`
	server := testGoogleRedeemBackend(p.Data(), redirect_uri, code, payload)
	defer server.Close()
	response, token, err := p.Redeem(redirect_uri, code)
	assert.Equal(t, nil, err)
	assert.Equal(t, payload, string(response))
	assert.Equal(t, "access_token refresh_token", token)
}

func TestGoogleProviderReedeemRefreshToken(t *testing.T) {
	p, refresh_token := newGoogleProvider(), "refresh_token"
	payload := `{"access_token":"new_access_token",` +
		` "expires_in":3920,` +
		` "token_type":"Bearer"}`
	server := testGoogleRedeemRefreshTokenBackend(p.Data(), refresh_token, payload)
	defer server.Close()
	token, err := p.redeemRefreshToken(refresh_token)
	assert.Equal(t, nil, err)
	assert.Equal(t, "new_access_token", token)
}

func TestGoogleProviderValidateToken(t *testing.T) {
	p := newGoogleProvider()
	access_token := "access_token"
	refresh_token := "refresh_token"
	full_token := access_token + " " + refresh_token

	server := testGoogleValidateTokenBackend(p.Data(), access_token, "")
	defer server.Close()

	ok, token := p.ValidateToken(full_token)
	assert.Equal(t, true, ok)
	assert.Equal(t, "", token)
}

func TestGoogleProviderValidateTokenReturnRefreshedToken(t *testing.T) {
	p := newGoogleProvider()
	access_token := "access_token"
	refresh_token := "refresh_token"
	full_token := access_token + " " + refresh_token

	// Not setting a path, etc. will force an error.
	validate_server := NewTestQueryBackend("", "", "")
	defer validate_server.Close()
	p.Data().ValidateUrl, _ = url.Parse(validate_server.URL)

	refresh_payload := `{"access_token":"new_access_token",` +
		` "expires_in":3920,` +
		` "token_type":"Bearer"}`
	refresh_server := testGoogleRedeemRefreshTokenBackend(
		p.Data(), refresh_token, refresh_payload)
	defer refresh_server.Close()

	ok, token := p.ValidateToken(full_token)
	assert.Equal(t, true, ok)
	assert.Equal(t, "new_access_token "+refresh_token, token)
}
