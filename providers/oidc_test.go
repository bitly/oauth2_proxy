package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	oidc "github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
)

func newJsonReturningRedeemServer(body []byte) (*url.URL, *httptest.Server) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header()["Content-Type"] = []string{"application/json"}
		rw.Write(body)
	}))
	u, _ := url.Parse(s.URL)
	return u, s
}

var issuer string = "https://exmple.org/"

type testVerifier struct {
}

func (t *testVerifier) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	parts := strings.Split(jwt, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt: %v", err)
	}
	return payload, nil
}

func newOidcProvider() *OIDCProvider {
	p := NewOIDCProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	p.Verifier = oidc.NewVerifier(issuer, &testVerifier{}, &oidc.Config{
		SkipClientIDCheck: true,
		SkipExpiryCheck:   true,
	})
	return p
}

// reusing redeemResponse from google_test

func TestOidcProviderLeavesUserlankByDefault(t *testing.T) {
	p := newOidcProvider()
	body, err := json.Marshal(redeemResponse{
		AccessToken:  "a1234",
		ExpiresIn:    10,
		RefreshToken: "refresh12345",
		IdToken:      base64.RawURLEncoding.EncodeToString([]byte(`{"typ": "jwt", "alg": "none"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("{\"iss\": \""+issuer+"\", \"email\": \"jane.doe@example.org\", \"email_verified\":true}")) + ".",
	})
	assert.Equal(t, nil, err)
	var server *httptest.Server
	p.RedeemURL, server = newJsonReturningRedeemServer(body)
	defer server.Close()

	session, err := p.Redeem("http://redirect/", "code1234")
	assert.Nil(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "", session.User)
}

func TestOidcProviderLeavesUserBlankIfConfiguredClaimIsMissing(t *testing.T) {
	p := newOidcProvider()
	p.ProviderData.UsernameClaim = "username"
	body, err := json.Marshal(redeemResponse{
		AccessToken:  "a1234",
		ExpiresIn:    10,
		RefreshToken: "refresh12345",
		IdToken:      base64.RawURLEncoding.EncodeToString([]byte(`{"typ": "jwt", "alg": "none"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("{\"iss\": \""+issuer+"\", \"email\": \"jane.doe@example.org\", \"email_verified\":true}")) + ".",
	})
	assert.Equal(t, nil, err)
	var server *httptest.Server
	p.RedeemURL, server = newJsonReturningRedeemServer(body)
	defer server.Close()

	session, err := p.Redeem("http://redirect/", "code1234")
	assert.Nil(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "", session.User)
}

func TestOidcProviderSetsUserFromConfiguredClaimIfPresent(t *testing.T) {
	p := newOidcProvider()
	p.ProviderData.UsernameClaim = "username"
	body, err := json.Marshal(redeemResponse{
		AccessToken:  "a1234",
		ExpiresIn:    10,
		RefreshToken: "refresh12345",
		IdToken:      base64.RawURLEncoding.EncodeToString([]byte(`{"typ": "jwt", "alg": "none"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("{\"iss\": \""+issuer+"\", \"email\": \"jane.doe@example.org\", \"email_verified\":true, \"username\": \"jd\"}")) + ".",
	})
	assert.Equal(t, nil, err)
	var server *httptest.Server
	p.RedeemURL, server = newJsonReturningRedeemServer(body)
	defer server.Close()

	session, err := p.Redeem("http://redirect/", "code1234")
	assert.Nil(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "jd", session.User)
}

func TestOidcProviderReturnsAnErrorIfConfiguredUsernameClaimIsNotStringValued(t *testing.T) {
	p := newOidcProvider()
	p.ProviderData.UsernameClaim = "username"
	body, err := json.Marshal(redeemResponse{
		AccessToken:  "a1234",
		ExpiresIn:    10,
		RefreshToken: "refresh12345",
		IdToken:      base64.RawURLEncoding.EncodeToString([]byte(`{"typ": "jwt", "alg": "none"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("{\"iss\": \""+issuer+"\", \"email\": \"jane.doe@example.org\", \"email_verified\":true, \"username\": true}")) + ".",
	})
	assert.Equal(t, nil, err)
	var server *httptest.Server
	p.RedeemURL, server = newJsonReturningRedeemServer(body)
	defer server.Close()

	_, err = p.Redeem("http://redirect/", "code1234")
	assert.NotNil(t, err)
}
