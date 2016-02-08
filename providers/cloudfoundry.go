package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type CloudFoundryProvider struct {
	*ProviderData
}

func NewCloudFoundryProvider(p *ProviderData) *CloudFoundryProvider {
	p.ProviderName = "CloudFoundry"
	// Defaults for Bosh-Lite, must be configured for other OAuth Endpoints.
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
                        Scheme: "http",
                        Host:   "login.bosh-lite.com",
                        Path:   "/oauth/authorize",
                }
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
                        Scheme: "http",
                        Host:   "login.bosh-lite.com",
                        Path:   "/oauth/token",
                }
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
                        Scheme: "http",
                        Host:   "login.bosh-lite.com",
                        Path:   "/userinfo",
                }
	}
	if p.Scope == "" {
		p.Scope = "openid"
	}
	return &CloudFoundryProvider{ProviderData: p}
}

func (p *CloudFoundryProvider) GetEmailAddress(s *SessionState) (string, error) {

	var email struct {
		Email   string `json:"email"`
	}

	params := url.Values{
		"access_token": {s.AccessToken},
	}
	endpoint := p.ValidateURL.String() + "?" + params.Encode()
	resp, err := http.DefaultClient.Get(endpoint)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got %d from %q %s", resp.StatusCode, endpoint, body)
	} else {
		log.Printf("got %d from %q %s", resp.StatusCode, endpoint, body)
	}

	if err := json.Unmarshal(body, &email); err != nil {
		return "", fmt.Errorf("%s unmarshaling %s", err, body)
	}

	return email.Email, nil
}
