package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type SimpleProvider struct {
	*ProviderData
}

func NewSimpleProvider(p *ProviderData) *SimpleProvider {
	p.ProviderName = "Simple"
	return &SimpleProvider{ProviderData: p}
}

func (p *SimpleProvider) GetEmailAddress(s *SessionState) (string, error) {
	_, email, err := p.getUserAndEmail(s)
	return email, err
}

func (p *SimpleProvider) GetUserName(s *SessionState) (string, error) {
	username, _, err := p.getUserAndEmail(s)
	if err != nil {
		return "", err
	}
	// Replace all spaces in the username by an alternative character because space is used to delimit session state.
	return strings.Replace(username, " ", "+", -1), err
}

func (p *SimpleProvider) getUserAndEmail(s *SessionState) (string, string, error) {
	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	req, err := http.NewRequest("GET", p.ProfileURL.String(), nil)
	if err != nil {
		return "", "", fmt.Errorf("could not create new GET request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("got %d from %q %s", resp.StatusCode, p.ProfileURL.String(), body)
	}

	log.Printf("got %d from %q %s", resp.StatusCode, p.ProfileURL.String(), body)

	if err := json.Unmarshal(body, &user); err != nil {
		return "", "", fmt.Errorf("%s unmarshaling %s", err, body)
	}

	return user.Username, user.Email, nil
}
