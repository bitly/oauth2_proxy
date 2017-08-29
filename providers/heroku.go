package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HerokuProvider struct {
	*ProviderData
}

func NewHerokuProvider(p *ProviderData) *HerokuProvider {
	const (
		idHost  = "id.heroku.com"
		apiHost = "api.heroku.com"
	)

	p.ProviderName = "Heroku"
	if p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{Scheme: "https",
			Host: idHost,
			Path: "/oauth/authorize"}
	}
	if p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{Scheme: "https",
			Host: idHost,
			Path: "/oauth/token"}
	}
	if p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{Scheme: "https",
			Host: idHost,
			Path: "/oauth/authorizations"}
	}
	if p.ProfileURL.String() == "" {
		p.ProfileURL = &url.URL{Scheme: "https",
			Host: apiHost,
			Path: "/account"}
	}
	if p.Scope == "" {
		p.Scope = "identity"
	}
	return &HerokuProvider{ProviderData: p}
}

func (p *HerokuProvider) GetEmailAddress(s *SessionState) (string, error) {
	req, _ := http.NewRequest("GET", p.ProfileURL.String(), nil)
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got %d from %q %s", resp.StatusCode, p.ProfileURL, body)
	}

	var account struct {
		Email string `json:"email"`
	}

	if err := json.Unmarshal(body, &account); err != nil {
		return "", fmt.Errorf("%s unmarshaling %s", err, body)
	}

	return account.Email, nil
}
