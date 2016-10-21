package providers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type GitLabProvider struct {
	*ProviderData
	Group string
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
			Path:   "/api/v4/user",
		}
	}
	if p.Scope == "" {
		p.Scope = "read_user"
	}
	return &GitLabProvider{ProviderData: p}
}

func (p *GitLabProvider) SetGroup(group string) {
	p.Group = group
}

func (p *GitLabProvider) hasGroup(accessToken string) (bool, error) {

	var groups []struct {
		Group string `json:"name"`
	}

	endpoint := p.ValidateURL.Scheme + "://" + p.ValidateURL.Host + "/api/v3/groups?access_token="+accessToken
	req, _ := http.NewRequest("GET", endpoint, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("got %d from %q %s", resp.StatusCode, endpoint, body)
	}

	if err := json.Unmarshal(body, &groups); err != nil {
		return false, err
	}

	for _, group := range groups {
		if( p.Group == group.Group) {
			// Found the group
			return true, nil
		}
	}

	log.Printf("Group %s not found in %s", p.Group, groups)
	return false, nil
}


func (p *GitLabProvider) GetEmailAddress(s *SessionState) (string, error) {

	// if we require a group, check that first
	if p.Group != "" {
		if ok, err := p.hasGroup(s.AccessToken); err != nil || !ok {
			return "", err
		}
	}

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
