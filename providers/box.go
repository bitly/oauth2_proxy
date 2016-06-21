package providers

import (
    "log"
    "net/http"
    "net/url"

    "github.com/bitly/oauth2_proxy/api"
)

type BoxProvider struct {
    *ProviderData
}

func NewBoxProvider(p *ProviderData) *BoxProvider {
    p.ProviderName = "Box"
    if p.LoginURL.String() == "" {
        p.LoginURL = &url.URL{Scheme: "https",
            Host: "account.box.com",
            Path: "/api/oauth2/authorize"}
    }
    if p.RedeemURL.String() == "" {
        p.RedeemURL = &url.URL{Scheme: "https",
            Host: "api.box.com",
            Path: "/oauth2/token"}
    }
    if p.ProfileURL.String() == "" {
        p.ProfileURL = &url.URL{Scheme: "https",
            Host: "api.box.com",
            Path: "/2.0/users/me"}
    }
    if p.ValidateURL.String() == "" {
        p.ValidateURL = p.ProfileURL
    }
    return &BoxProvider{ProviderData: p}
}

func (p *BoxProvider) GetEmailAddress(s *SessionState) (string, error) {
    req, err := http.NewRequest("GET", p.ProfileURL.String(), nil)
    if err != nil {
        log.Printf("failed building request %s", err)
        return "", err
    }
    req.Header.Set("Authorization:", "Bearer "+s.AccessToken)
    json, err := api.Request(req)
    if err != nil {
        log.Printf("failed making request %s", err)
        return "", err
    }
    return json.Get("login").String()
}
