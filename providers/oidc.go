package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	oidc "github.com/coreos/go-oidc"

	"github.com/bitly/oauth2_proxy/api"
)

type OIDCProvider struct {
	*ProviderData

	Verifier *oidc.IDTokenVerifier
}

func NewOIDCProvider(p *ProviderData) *OIDCProvider {
	p.ProviderName = "OpenID Connect"
	return &OIDCProvider{ProviderData: p}
}

func (p *OIDCProvider) Redeem(redirectURL, code string) (s *SessionState, err error) {
	ctx := context.Background()
	c := oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: p.RedeemURL.String(),
		},
		RedirectURL: redirectURL,
	}
	token, err := c.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %v", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("token response did not contain an id_token")
	}

	// Parse and verify ID Token payload.
	idToken, err := p.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("could not verify id_token: %v", err)
	}

	// Extract custom claims.
	var claims struct {
		Email    string `json:"email"`
		Verified *bool  `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse id_token claims: %v", err)
	}

	if claims.Email == "" {
		if p.ProfileURL.String() == "" {
			return nil, fmt.Errorf("id_token did not contain an email")
		}

		// If the userinfo endpoint profileURL is defined, then there
		// is a chance the userinfo contents at the profileURL contain
		// the email.  Make a query to the userinfo endpoint, and
		// attempt to locate the email from there.

		req, err := http.NewRequest("GET", p.ProfileURL.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header = getOIDCHeader(token.AccessToken)

		json, err := api.Request(req)
		if err != nil {
			return nil, err
		}

		email, err := json.Get("email").String()
		if err != nil {
			return nil, fmt.Errorf("id_token nor userinfo endpoint did not contain an email")
		}

		claims.Email = email
	}
	if claims.Verified != nil && !*claims.Verified {
		return nil, fmt.Errorf("id_token failed verification")
	}

	s = &SessionState{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresOn:    token.Expiry,
		Email:        claims.Email,
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

func getOIDCHeader(access_token string) http.Header {
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	return header
}
