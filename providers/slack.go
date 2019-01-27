package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type SlackProvider struct {
	*ProviderData
	TeamID string
}

// Slack API Response for https://api.slack.com/methods/users.identity
type SlackUserIdentityResponse struct {
	OK   bool
	User SlackUserItem
	Team SlackUserItem
}

type SlackUserItem struct {
	ID    string
	Name  string
	Email string
}

func NewSlackProvider(p *ProviderData) *SlackProvider {
	p.ProviderName = "slack"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/oauth/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/api/oauth.access",
		}
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/api",
		}
	}
	if p.Scope == "" {
		p.Scope = "identity.basic identity.email"
	}
	return &SlackProvider{ProviderData: p}
}

func (p *SlackProvider) SetTeamID(team string) {
	p.TeamID = team
	// If a team id is set we can restrict login to this team directly at login
	params, _ := url.ParseQuery(p.LoginURL.RawQuery)
	params.Set("team", team)
	p.LoginURL.RawQuery = params.Encode()
}

func (p *SlackProvider) getIdentity(accessToken string) (*SlackUserIdentityResponse, error) {
	params := url.Values{
		"token": {accessToken},
	}
	endpoint := &url.URL{
		Scheme:   p.ValidateURL.Scheme,
		Host:     p.ValidateURL.Host,
		Path:     path.Join(p.ValidateURL.Path, "/users.identity"),
		RawQuery: params.Encode(),
	}
	req, _ := http.NewRequest("GET", endpoint.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"got %d from %q %s", resp.StatusCode, endpoint.String(), body)
	}

	var userIdentity SlackUserIdentityResponse
	if err := json.Unmarshal(body, &userIdentity); err != nil {
		return nil, err
	}

	if userIdentity.OK == true {
		return &userIdentity, nil
	}
	return nil, fmt.Errorf("slack response is not ok: %v", userIdentity)
}

func (p *SlackProvider) hasTeamID(resp *SlackUserIdentityResponse) (bool, error) {
	if resp.Team.ID != "" {
		return resp.Team.ID == p.TeamID, nil
	}

	return false, fmt.Errorf("no team id found")
}

func (p *SlackProvider) GetEmailAddress(s *SessionState) (string, error) {
	userIdentity, err := p.getIdentity(s.AccessToken)
	if err != nil {
		return "", nil
	}

	// if we require a TeamId, check that first
	if p.TeamID != "" {
		if ok, err := p.hasTeamID(userIdentity); err != nil || !ok {
			return "", err
		}
	}

	if email := userIdentity.User.Email; email != "" {
		return email, nil
	}

	return "", nil
}
