package providers

import (
	"github.com/bitly/oauth2_proxy/cookie"
)

type Provider interface {
	Data() *ProviderData
	GetEmailAddress(*SessionState) (string, error)
	Redeem(string, string) (*SessionState, error)
	ValidateGroup(string) bool
	ValidateSessionState(*SessionState) bool
	GetLoginURL(redirectURI, finalRedirect string) string
	RefreshSessionIfNeeded(*SessionState) (bool, error)
	SessionFromCookie(string, *cookie.Cipher) (*SessionState, error)
	CookieForSession(*SessionState, *cookie.Cipher) (string, error)
}

func New(provider string, p *ProviderData) Provider {
	switch provider {
	case "myusa":
		return NewMyUsaProvider(p)
	case "linkedin":
		return NewLinkedInProvider(p)
	case "github":
		return NewGitHubProvider(p)
	case "oidc":
		return NewOIDCProvider(p)
	case "google":
		return NewGoogleProvider(p)
	default:
		// TODO(philips): this should error out
		return NewGoogleProvider(p)
	}
}
