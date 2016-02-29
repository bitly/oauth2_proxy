package providers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/oidc"
)

type OIDCProvider struct {
	*ProviderData
	clientConfig     oidc.ClientConfig
	RedeemRefreshURL *url.URL
}

func NewOIDCProvider(p *ProviderData) *OIDCProvider {
	var err error

	p.ProviderName = "OpenID Connect"

	cc := oidc.ClientCredentials{
		ID:     p.ClientID,
		Secret: p.ClientSecret,
	}

	var tlsConfig tls.Config
	// TODO: do handling of custom certs
	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tlsConfig}}

	var cfg oidc.ProviderConfig
	for {
		cfg, err = oidc.FetchProviderConfig(httpClient, p.DiscoveryURL.String())
		if err == nil {
			break
		}

		sleep := 3 * time.Second
		log.Printf("Failed fetching provider config, trying again in %v: %v", sleep, err)
		time.Sleep(sleep)
	}

	u, err := url.Parse(cfg.TokenEndpoint)
	if err != nil {
		panic(err)
	}
	p.ValidateURL = u
	u, err = url.Parse(cfg.AuthEndpoint)
	if err != nil {
		panic(err)
	}
	p.RedeemURL = u
	p.Scope = "email"

	ccfg := oidc.ClientConfig{
		HTTPClient:     httpClient,
		ProviderConfig: cfg,
		Credentials:    cc,
	}

	client, err := oidc.NewClient(ccfg)
	if err != nil {
		log.Fatalf("Unable to create Client: %v", err)
	}

	client.SyncProviderConfig(p.DiscoveryURL.String())

	oac, err := client.OAuthClient()
	if err != nil {
		panic("unable to proceed")
	}

	login, err := url.Parse(oac.AuthCodeURL("", "", ""))
	if err != nil {
		panic("unable to proceed")
	}

	p.LoginURL = login

	return &OIDCProvider{
		ProviderData: p,
		clientConfig: ccfg,
	}
}

func (p *OIDCProvider) Redeem(redirectURL, code string) (s *SessionState, err error) {
	c, err := oidc.NewClient(p.clientConfig)
	if err != nil {
		log.Fatalf("Unable to create Client: %v", err)
	}

	tok, err := c.ExchangeAuthCode(code)
	if err != nil {
		log.Printf("exchange auth error: %v\n", err)
		return nil, err
	}

	claims, err := tok.Claims()
	if err != nil {
		log.Printf("token claims error: %v", err)
		return nil, err
	}

	s = &SessionState{
		AccessToken:  tok.Data(),
		RefreshToken: tok.Data(),
		ExpiresOn:    time.Now().Add(time.Duration(claims["exp"].(float64)) * time.Second).Truncate(time.Second),
		Email:        claims["email"].(string),
	}

	return
}

func (p *OIDCProvider) RefreshSessionIfNeeded(s *SessionState) (bool, error) {
	if s == nil || s.ExpiresOn.After(time.Now()) || s.RefreshToken == "" {
		return false, nil
	}

	origExpiration := s.ExpiresOn
	s.ExpiresOn = time.Now().Add(time.Second).Truncate(time.Second)
	fmt.Printf("refreshed access token %s (expired on %s)\n", s, origExpiration)
	return false, nil
}
