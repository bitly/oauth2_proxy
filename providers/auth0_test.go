package providers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testAuth0Provider() *Auth0Provider {
	p := NewAuth0Provider(
		&ProviderData{
			ApprovalPrompt:    "approvalPrompt",
			ClientID:          "clientID",
			LoginURL:          &url.URL{},
			RedeemURL:         &url.URL{},
			ProfileURL:        &url.URL{},
			ValidateURL:       &url.URL{},
			ProtectedResource: &url.URL{},
		})
	return p
}

func TestAuth0ProviderDefaultsConfigure(t *testing.T) {
	p := testAuth0Provider()
	p.Configure("domain.zone.auth0.com")

	assert.Equal(t, "Auth0", p.Data().ProviderName)
	assert.Equal(t, "domain.zone.auth0.com", p.Domain)
	assert.Equal(t, "https://domain.zone.auth0.com/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://domain.zone.auth0.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://domain.zone.auth0.com/userinfo",
		p.Data().ProfileURL.String())
	assert.Equal(t, "",
		p.Data().ProtectedResource.String())
	assert.Equal(t, "https://domain.zone.auth0.com/userinfo",
		p.Data().ValidateURL.String())
	assert.Equal(t, "openid profile email", p.Data().Scope)
}

func TestAuth0GetLoginUrl(t *testing.T) {
	p := testAuth0Provider()
	p.Configure("domain.zone.auth0.com")
	actual := p.GetLoginURL("http://localhost:1234", "state")
	assert.Equal(t, "https://domain.zone.auth0.com/authorize?"+
		"approval_prompt=approvalPrompt"+
		"&client_id=clientID"+
		"&redirect_uri=http%3A%2F%2Flocalhost%3A1234"+
		"&response_type=code"+
		"&scope=openid+profile+email"+
		"&state=state",
		actual)
}
