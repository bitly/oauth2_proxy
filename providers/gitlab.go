package providers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

type GitLabProvider struct {
	*ProviderData
}

func NewGitLabProvider(p *ProviderData) *GitLabProvider {
	p.ProviderName = "GitLab"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "gitlab.com",
			Path:   "/oauth/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "gitlab.com",
			Path:   "/oauth/token",
		}
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "gitlab.com",
			Path:   "/api/v3/user",
		}
	}
	if p.Scope == "" {
		p.Scope = "read_user"
	}
	return &GitLabProvider{ProviderData: p}
}

func (p *GitLabProvider) GetEmailAddress(s *SessionState) (string, error) {
	req, err := http.NewRequest("GET",
		p.ValidateURL.String()+"?access_token="+s.AccessToken, nil)
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

func (p *GitLabProvider) IsTwoFactorAuthEnabled(s *SessionState) (hasTwoFactor bool, redirectTo string, err error) {
	redirectTo = fmt.Sprintf("%v://%v/profile/two_factor_auth", p.LoginURL.Scheme, p.LoginURL.Host)

	req, err := http.NewRequest("GET",
		p.ValidateURL.String()+"?access_token="+s.AccessToken, nil)
	if err != nil {
		log.Printf("gitlab.IsTwoFactorAuthEnabled: failed building request %s", err)
		return false, redirectTo, err
	}
	json, err := api.Request(req)
	log.Printf("gitlab.IsTwoFactorAuthEnabled: received response: %v", json)
	if err != nil {
		log.Printf("failed making request %s", err)
		return false, redirectTo, err
	}

	hasTwoFactor, err = json.Get("two_factor_enabled").Bool()
	return
}
