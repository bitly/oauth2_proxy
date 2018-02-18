package providers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/brunnels/oauth2_proxy/api"
)

type OwncloudProvider struct {
	*ProviderData
}

func NewOwncloudProvider(p *ProviderData) *OwncloudProvider {
	p.ProviderName = "Owncloud"

	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/index.php/apps/oauth2/authorize",
		}
	}

	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/index.php/apps/oauth2/api/v1/token",
		}
	}

	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/ocs/v1.php/cloud/user",
		}
	}
	return &OwncloudProvider{ProviderData: p}
}

func getOwncloudHeader(access_token string) http.Header {
	header := make(http.Header)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	return header
}

func (p *OwncloudProvider) GetEmailAddress(s *SessionState) (string, error) {
	if s.AccessToken == "" {
		return "", errors.New("missing access token")
	}

	req, err := http.NewRequest("GET", p.ValidateURL.String()+"?format=json", nil)

	if err != nil {
		return "", err
	}
	req.Header = getOwncloudHeader(s.AccessToken)

	type result struct {
		Ocs struct {
			Meta struct {
				Status     string `json:"status"`
				StatusCode int    `json:"statuscode"`
				Message    string `json:"message"`
			} `json:"meta"`
			Data struct {
				Id          string `json:"id"`
				DisplayName string `json:"display-name"`
				Email       string `json:"email"`
			} `json:"data"`
		} `json:"ocs"`
	}
	var r result
	err = api.RequestJson(req, &r)
	if err != nil {
		return "", err
	}
	if r.Ocs.Data.Id == "" {
		return "", errors.New("no id")
	}
	return r.Ocs.Data.Id + "@" + p.LoginURL.Host, nil
}