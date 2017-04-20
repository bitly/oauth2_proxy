package providers

import (
	"errors"
	"fmt"
	"github.com/bitly/oauth2_proxy/api"
	"log"
	"net/http"
	"net/url"
)

type ZendeskProvider struct {
	*ProviderData
	Subdomain string
}

func NewZendeskProvider(p *ProviderData) *ZendeskProvider {
	p.ProviderName = "Zendesk"

	if p.Scope == "" {
		p.Scope = "read"
	}

	return &ZendeskProvider{ProviderData: p}
}

func (p *ZendeskProvider) Configure(subdomain string) {
	p.Subdomain = subdomain

	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   p.Subdomain + ".zendesk.com",
			Path:   "/oauth/authorizations/new"}
	}

	if p.ProfileURL == nil || p.ProfileURL.String() == "" {
		p.ProfileURL = &url.URL{
			Scheme:   "https",
			Host:     p.Subdomain + ".zendesk.com",
			Path:     "/api/v2/users/me.json"}
	}

	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   p.Subdomain + ".zendesk.com",
			Path:   "/oauth/tokens"}
	}

	if p.ProtectedResource == nil || p.ProtectedResource.String() == "" {
		p.ProtectedResource = &url.URL{
			Scheme: "https",
			Host:   p.Subdomain + ".zendesk.com",
		}
	}
}

func getZendeskHeader(access_token string) http.Header {
	header := make(http.Header)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	return header
}

func (p *ZendeskProvider) GetEmailAddress(s *SessionState) (string, error) {
	if s.AccessToken == "" {
		return "", errors.New("missing access token")
	}
	req, err := http.NewRequest("GET", p.ProfileURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header = getAzureHeader(s.AccessToken)

	json, err := api.Request(req)

	if err != nil {
		log.Printf("failed making request %s", err)
		return "", err
	}

	email, err := json.Get("user").Get("email").String()
	if err != nil {
		fmt.Printf("failed parsing JSON '%s'; error %s", email, err)
		return "", err
	}

	return email, nil
}
