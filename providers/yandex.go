package providers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

type YandexProvider struct {
	*ProviderData
}

func NewYandexProvider(p *ProviderData) *YandexProvider {
	p.ProviderName = "Yandex"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "oauth.yandex.com",
			Path:   "/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "oauth.yandex.com",
			Path:   "/token",
		}
	}
	if p.ProfileURL == nil || p.ProfileURL.String() == "" {
		p.ProfileURL = &url.URL{
			Scheme: "https",
			Host:   "login.yandex.ru",
			Path:   "/info",
		}
	}
	if p.Scope == "" {
		p.Scope = "login:email"
	}
	return &YandexProvider{ProviderData: p}
}

func (p *YandexProvider) GetEmailAddress(s *SessionState) (string, error) {
	req, err := http.NewRequest("GET",
		p.ProfileURL.String()+"?oauth_token="+s.AccessToken, nil)
	if err != nil {
		log.Printf("failed building request %s", err)
		return "", err
	}
	json, err := api.Request(req)
	if err != nil {
		log.Printf("failed making request %s", err)
		return "", err
	}
	return json.Get("default_email").String()
}
