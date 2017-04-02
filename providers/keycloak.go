package providers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

type KeycloakProvider struct {
	*ProviderData
}

func NewKeycloakProvider(p *ProviderData) *KeycloakProvider {
	p.ProviderName = "Keycloak"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "keycloak.org",
			Path:   "/oauth/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "keycloak.org",
			Path:   "/oauth/token",
		}
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "keycloak.org",
			Path:   "/api/v3/user",
		}
	}
	if p.Scope == "" {
		p.Scope = "api"
	}
	return &KeycloakProvider{ProviderData: p}
}

func (p *KeycloakProvider) GetEmailAddress(s *SessionState) (string, error) {

	req, err := http.NewRequest("GET", p.ValidateURL.String(), nil)
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	if err != nil {
		log.Printf("failed building request %s", err)
		return "", err
	}
	json, err := api.Request(req)
	if err != nil {
		log.Printf("failed making request %s", err)
		return "", err
	}
	return json.Get("email").String()
}
