package providers

import (
	"github.com/bitly/oauth2_proxy/cookie"
)

// Provider is the primary interface for an authentication provider
// all provider
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

// RoleProvider is an optional interface that exposes a list of roles
// for a user. For Providers like GitHub this would be the teams the user
// is a member of.
type RoleProvider interface {
	GetUserRoles() string
	SetUserRoles(string) (bool, error)
}

// New gives you an instance of the given provider
func New(provider string, p *ProviderData) Provider {
	switch provider {
	case "myusa":
		return NewMyUsaProvider(p)
	case "linkedin":
		return NewLinkedInProvider(p)
	case "facebook":
		return NewFacebookProvider(p)
	case "github":
		return NewGitHubProvider(p)
	case "azure":
		return NewAzureProvider(p)
	case "gitlab":
		return NewGitLabProvider(p)
	default:
		return NewGoogleProvider(p)
	}
}
