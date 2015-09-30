package main

import (
	"crypto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func testOptions() *Options {
	o := NewOptions()
	o.Upstreams = append(o.Upstreams, "http://127.0.0.1:8080/")
	o.CookieSecret = "foobar"
	o.ClientID = "bazquux"
	o.ClientSecret = "xyzzyplugh"
	o.EmailDomains = []string{"*"}
	return o
}

func errorMsg(msgs []string) string {
	result := make([]string, 0)
	result = append(result, "Invalid configuration:")
	result = append(result, msgs...)
	return strings.Join(result, "\n  ")
}

func TestNewOptions(t *testing.T) {
	o := NewOptions()
	o.EmailDomains = []string{"*"}
	err := o.Validate()
	assert.NotEqual(t, nil, err)

	expected := errorMsg([]string{
		"missing setting: upstream",
		"missing setting: cookie-secret",
		"missing setting: client-id",
		"missing setting: client-secret"})
	assert.Equal(t, expected, err.Error())
}

func TestGoogleGroupOptions(t *testing.T) {
	o := testOptions()
	o.GoogleGroups = []string{"googlegroup"}
	err := o.Validate()
	assert.NotEqual(t, nil, err)

	expected := errorMsg([]string{
		"missing setting: google-admin-email",
		"missing setting: google-service-account-json"})
	assert.Equal(t, expected, err.Error())
}

func TestGoogleGroupInvalidFile(t *testing.T) {
	o := testOptions()
	o.GoogleGroups = []string{"test_group"}
	o.GoogleAdminEmail = "admin@example.com"
	o.GoogleServiceAccountJSON = "file_doesnt_exist.json"
	err := o.Validate()
	assert.NotEqual(t, nil, err)

	expected := errorMsg([]string{
		"invalid Google credentials file: file_doesnt_exist.json",
	})
	assert.Equal(t, expected, err.Error())
}

func TestInitializedOptions(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())
}

// Note that it's not worth testing nonparseable URLs, since url.Parse()
// seems to parse damn near anything.
func TestRedirectURL(t *testing.T) {
	o := testOptions()
	o.RedirectURL = "https://myhost.com/oauth2/callback"
	assert.Equal(t, nil, o.Validate())
	expected := &url.URL{
		Scheme: "https", Host: "myhost.com", Path: "/oauth2/callback"}
	assert.Equal(t, expected, o.redirectURL)
}

func TestProxyURLs(t *testing.T) {
	o := testOptions()
	o.Upstreams = append(o.Upstreams, "http://127.0.0.1:8081")
	assert.Equal(t, nil, o.Validate())
	expected := []*url.URL{
		&url.URL{Scheme: "http", Host: "127.0.0.1:8080", Path: "/"},
		// note the '/' was added
		&url.URL{Scheme: "http", Host: "127.0.0.1:8081", Path: "/"},
	}
	assert.Equal(t, expected, o.proxyURLs)
}

func TestCompiledRegex(t *testing.T) {
	o := testOptions()
	regexps := []string{"/foo/.*", "/ba[rz]/quux"}
	o.SkipAuthRegex = regexps
	assert.Equal(t, nil, o.Validate())
	actual := make([]string, 0)
	for _, regex := range o.CompiledRegex {
		actual = append(actual, regex.String())
	}
	assert.Equal(t, regexps, actual)
}

func TestCompiledRegexError(t *testing.T) {
	o := testOptions()
	o.SkipAuthRegex = []string{"(foobaz", "barquux)"}
	err := o.Validate()
	assert.NotEqual(t, nil, err)

	expected := errorMsg([]string{
		"error compiling regex=\"(foobaz\" error parsing regexp: " +
			"missing closing ): `(foobaz`",
		"error compiling regex=\"barquux)\" error parsing regexp: " +
			"unexpected ): `barquux)`"})
	assert.Equal(t, expected, err.Error())
}

func TestDefaultProviderApiSettings(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())
	p := o.provider.Data()
	assert.Equal(t, "https://accounts.google.com/o/oauth2/auth?access_type=offline",
		p.LoginURL.String())
	assert.Equal(t, "https://www.googleapis.com/oauth2/v3/token",
		p.RedeemURL.String())
	assert.Equal(t, "", p.ProfileURL.String())
	assert.Equal(t, "profile email", p.Scope)
}

func TestPassAccessTokenRequiresSpecificCookieSecretLengths(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())

	assert.Equal(t, false, o.PassAccessToken)
	o.PassAccessToken = true
	o.CookieSecret = "cookie of invalid length-"
	assert.NotEqual(t, nil, o.Validate())

	o.PassAccessToken = false
	o.CookieRefresh = time.Duration(24) * time.Hour
	assert.NotEqual(t, nil, o.Validate())

	o.CookieSecret = "16 bytes AES-128"
	assert.Equal(t, nil, o.Validate())

	o.CookieSecret = "24 byte secret AES-192--"
	assert.Equal(t, nil, o.Validate())

	o.CookieSecret = "32 byte secret for AES-256------"
	assert.Equal(t, nil, o.Validate())
}

func TestCookieRefreshMustBeLessThanCookieExpire(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())

	o.CookieSecret = "0123456789abcdef"
	o.CookieRefresh = o.CookieExpire
	assert.NotEqual(t, nil, o.Validate())

	o.CookieRefresh -= time.Duration(1)
	assert.Equal(t, nil, o.Validate())
}

func TestValidateUpstreamSignatureKeys(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())

	o.Upstreams = []string{
		"https://foo.com:8000",
		"https://bar.com/bar",
		"https://baz.com",
	}
	o.SignatureKey = "sha1:default secret"
	o.UpstreamKeys = []string{
		"foo.com:8000=sha1:secret0",
		"bar.com=sha1:secret1",
		"baz.com=sha1:secret2",
	}

	assert.Equal(t, nil, o.Validate())
	assert.Equal(t, o.upstreamKeys, map[string]*SignatureData{
		"foo.com:8000": &SignatureData{crypto.SHA1, "secret0"},
		"bar.com":      &SignatureData{crypto.SHA1, "secret1"},
		"baz.com":      &SignatureData{crypto.SHA1, "secret2"},
	})
}

func TestValidateUpstreamSignatureKeysWithErrors(t *testing.T) {
	o := testOptions()
	assert.Equal(t, nil, o.Validate())

	o.Upstreams = []string{
		"https://bar.com/bar",
		"https://baz.com",
		"https://quux.com",
		"https://xyzzy.com",
	}
	o.SignatureKey = "unsupported:default secret"
	o.UpstreamKeys = []string{
		"foo.com:8000=sha1:secret0",
		"bar.com=secret1",
		"baz.com:sha1:secret2",
		"quux.com=sha1:secret3",
		"quux.com=sha1:secret4",
		"xyzzy.com=unsupported:secret5",
	}

	err := o.Validate()
	assert.NotEqual(t, nil, err)
	expected := errorMsg([]string{
		"unsupported signature hash algorithm: " +
			"unsupported:default secret",
		"invalid upstream key specs:",
		"  invalid signature hash:key spec: bar.com=secret1",
		"  baz.com:sha1:secret2",
		"  unsupported signature hash algorithm: " +
			"xyzzy.com=unsupported:secret5",
		"specs with hosts that do not match any defined upstreams:",
		"  foo.com:8000=sha1:secret0",
		"specs that duplicate other host specs:",
		"  quux.com=sha1:secret4",
	})
	assert.Equal(t, err.Error(), expected)
}
