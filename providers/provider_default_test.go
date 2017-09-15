package providers

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestRefresh(t *testing.T) {
	p := &ProviderData{}
	refreshed, err := p.RefreshSessionIfNeeded(&SessionState{
		ExpiresOn: time.Now().Add(time.Duration(-11) * time.Minute),
	})
	assert.Equal(t, false, refreshed)
	assert.Equal(t, nil, err)
}

func TestDefaultProviderRedeem(t *testing.T) {
	clientID, clientSecret, scope := "abc", "def", "profile"
	loginURL, redeemURL, profileURL, validateURL := &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/oauth/auth",
	}, &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/oauth/token",
	}, &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/oauth/profile",
	}, &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/oauth/tokeninfo",
	}

	p := ProviderData{
		HTTPBasicAuth:     true,
		ProviderName:      "Tester",
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		LoginURL:          loginURL,
		RedeemURL:         redeemURL,
		ProfileURL:        profileURL,
		ProtectedResource: nil,
		ValidateURL:       validateURL,
		Scope:             scope,
		ApprovalPrompt:    "prompt",
	}
	body, err := json.Marshal(redeemResponse{
		AccessToken:  "a1234",
		ExpiresIn:    10,
		RefreshToken: "refresh12345",
		IdToken:      "ignored prefix." + base64.URLEncoding.EncodeToString([]byte(`{"email": "michael.bland@gsa.gov", "email_verified":true}`)),
	})
	assert.Equal(t, nil, err)
	var server *httptest.Server
	var req map[string]interface{}
	p.RedeemURL, server, req = newRedeemServer(body)
	defer server.Close()

	t.Run("HTTPBasicAuth yes", func(t *testing.T) {
		p.HTTPBasicAuth = true
		_, err := p.Redeem("http://redirect/", "code1234")
		assert.Equal(t, nil, err)
		assert.Equal(t, p.ClientID, req["username"])
		assert.Equal(t, p.ClientSecret, req["password"])

		reqbody, err := url.ParseQuery(string(req["body"].([]byte)))
		assert.Equal(t, nil, err)
		assert.Equal(t, "", reqbody.Get("client_secret"))
		assert.Equal(t, p.ClientID, reqbody.Get("client_id"))
	})
	t.Run("HTTPBasicAuth no", func(t *testing.T) {
		p.HTTPBasicAuth = false
		_, err := p.Redeem("http://redirect/", "code1234")
		assert.Equal(t, nil, err)
		assert.Equal(t, "", req["username"])
		assert.Equal(t, "", req["password"])

		reqbody, err := url.ParseQuery(string(req["body"].([]byte)))
		assert.Equal(t, nil, err)
		assert.Equal(t, p.ClientID, reqbody.Get("client_id"))
		assert.Equal(t, p.ClientSecret, reqbody.Get("client_secret"))
	})
}

/*
	LEFT:
	1- WithBasicAuth: check that did *not* set client_secret in body
	2- WithoutBasicAuth: check that *did* set client_secret in body
	3- WithoutBasicAuth: check that did *not* set httpbasicauth
	4- use the multiple test wrapper
*/
