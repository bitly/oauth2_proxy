package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testOpenshiftProvider(hostname string, project string) *OpenshiftProvider {
	p := NewOpenshiftProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        "",
		})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}
	if project != "" {
		p.SetProject(project)
	}
	return p
}

func testOpenshiftBackend(payload string) *httptest.Server {

	validPaths := map[string]bool{
		"/oapi/v1/users/~":  true,
		"/oapi/v1/projects": true,
	}

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			if !validPaths[url.Path] {
				w.WriteHeader(404)
			} else if r.Header.Get("Authorization") != "Bearer imaginary_access_token" {
				w.WriteHeader(403)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}

func TestOpenshiftProviderDefaults(t *testing.T) {
	p := testOpenshiftProvider("", "")
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Openshift", p.Data().ProviderName)
	assert.Equal(t, "",
		p.Data().LoginURL.String())
	assert.Equal(t, "",
		p.Data().RedeemURL.String())
	assert.Equal(t, "",
		p.Data().ValidateURL.String())
	assert.Equal(t, "",
		p.Project)
	assert.Equal(t, "user:info user:check-access", p.Data().Scope)
}

func TestOpenshiftProviderOverrides(t *testing.T) {
	p := NewOpenshiftProvider(
		&ProviderData{
			LoginURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/authorize"},
			RedeemURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/token"},
			ValidateURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oapi/v1"},
			Scope: "user:info user:check-access"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Openshift", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/oapi/v1",
		p.Data().ValidateURL.String())
	assert.Equal(t, "user:info user:check-access", p.Data().Scope)
}

func TestOpenshiftProjectScope(t *testing.T) {
	p := testOpenshiftProvider("", "default")
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "user:info user:check-access user:list-projects", p.Data().Scope)
}

func TestOpenshiftProviderGetEmailAddress(t *testing.T) {
	b := testOpenshiftBackend(`{"kind":"User","apiVersion":"v1","metadata":{"name":"user@example.com","selfLink":"/oapi/v1/users/user%40example.com","uid":"4f85ea38-d35a-46f5-8682-1d7dc424d948","resourceVersion":"75260781","creationTimestamp":"2017-03-23T10:45:49Z"},"identities":["Company-AD:user@example.com"],"groups":[]}`)
	defer b.Close()
	b_url, _ := url.Parse(b.URL)
	p := testOpenshiftProvider(b_url.Host, "")

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user@example.com", email)
}

func TestOpenshiftProviderGetEmailAddressFailedRequest(t *testing.T) {
	b := testOpenshiftBackend("unused payload")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testOpenshiftProvider(b_url.Host, "")

	// We'll trigger a request failure by using an unexpected access
	// token. Alternatively, we could allow the parsing of the payload as
	// JSON to fail.
	session := &SessionState{AccessToken: "unexpected_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}

func TestOpenshiftProviderGetEmailAddressEmailNotPresentInPayload(t *testing.T) {
	b := testOpenshiftBackend("{\"foo\": \"bar\"}")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testOpenshiftProvider(b_url.Host, "")

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}

func TestOpenshiftProjectAllowAccess(t *testing.T) {
	b := testOpenshiftBackend(`{"kind":"ProjectList","apiVersion":"v1","metadata":{"selfLink":"/oapi/v1/projects"},"items":[{"metadata":{"name":"default","selfLink":"/oapi/v1/projects/default","uid":"a01fa48a-b265-11e7-b662-fa163e911gc5","resourceVersion":"81632269","creationTimestamp":"2017-10-16T11:32:00Z","annotations":{"openshift.io/description":"","openshift.io/display-name":"","openshift.io/requester":"user@example.com","openshift.io/sa.scc.mcs":"s0:c30,c10","openshift.io/sa.scc.supplemental-groups":"1000690000/10000","openshift.io/sa.scc.uid-range":"1000690000/10000"}},"spec":{"finalizers":["openshift.io/origin","kubernetes"]},"status":{"phase":"Active"}}]}`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testOpenshiftProvider(b_url.Host, "default")
	t.Log(p.Project)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	access, err := p.hasProject(session.AccessToken)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, access)
}

func TestOpenshiftProjectDenyAccess(t *testing.T) {
	b := testOpenshiftBackend(`{"kind":"ProjectList","apiVersion":"v1","metadata":{"selfLink":"/oapi/v1/projects"},"items":[{"metadata":{"name":"default","selfLink":"/oapi/v1/projects/default","uid":"a01fa48a-b265-11e7-b662-fa163e911gc5","resourceVersion":"81632269","creationTimestamp":"2017-10-16T11:32:00Z","annotations":{"openshift.io/description":"","openshift.io/display-name":"","openshift.io/requester":"user@example.com","openshift.io/sa.scc.mcs":"s0:c30,c10","openshift.io/sa.scc.supplemental-groups":"1000690000/10000","openshift.io/sa.scc.uid-range":"1000690000/10000"}},"spec":{"finalizers":["openshift.io/origin","kubernetes"]},"status":{"phase":"Active"}}]}`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testOpenshiftProvider(b_url.Host, "imaginaryProject")

	session := &SessionState{AccessToken: "imaginary_access_token"}
	access, err := p.hasProject(session.AccessToken)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, false, access)
}
