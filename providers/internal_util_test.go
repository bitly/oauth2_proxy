package providers

import (
	"github.com/bitly/go-simplejson"
	"github.com/bmizerany/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type ValidateTokenTestProvider struct {
	*ProviderData
}

func (tp *ValidateTokenTestProvider) GetEmailAddress(
	unused_auth_response *simplejson.Json,
	unused_access_token string) (string, error) {
	return "", nil
}

// Note that we're testing the internal validateToken() used to implement
// several Provider's ValidateToken() implementations
func (tp *ValidateTokenTestProvider) ValidateToken(access_token string) bool {
	return false
}

type ValidateTokenTest struct {
	backend       *httptest.Server
	response_code int
	provider      *ValidateTokenTestProvider
}

func NewValidateTokenTest() *ValidateTokenTest {
	var vt_test ValidateTokenTest

	vt_test.backend = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oauth/tokeninfo":
				w.WriteHeader(vt_test.response_code)
				w.Write([]byte("only code matters; contents disregarded"))
			default:
				w.WriteHeader(500)
				w.Write([]byte("unknown URL"))
			}
		}))
	backend_url, _ := url.Parse(vt_test.backend.URL)
	vt_test.provider = &ValidateTokenTestProvider{
		ProviderData: &ProviderData{
			ValidateUrl: &url.URL{
				Scheme: "http",
				Host:   backend_url.Host,
				Path:   "/oauth/tokeninfo",
			},
		},
	}
	vt_test.response_code = 200
	return &vt_test
}

func (vt_test *ValidateTokenTest) Close() {
	vt_test.backend.Close()
}

func TestValidateTokenValidToken(t *testing.T) {
	vt_test := NewValidateTokenTest()
	defer vt_test.Close()
	assert.Equal(t, true, validateToken(vt_test.provider, "foobar"))
}

func TestValidateTokenEmptyToken(t *testing.T) {
	vt_test := NewValidateTokenTest()
	defer vt_test.Close()
	assert.Equal(t, false, validateToken(vt_test.provider, ""))
}

func TestValidateTokenEmptyValidateUrl(t *testing.T) {
	vt_test := NewValidateTokenTest()
	defer vt_test.Close()
	vt_test.provider.Data().ValidateUrl = nil
	assert.Equal(t, false, validateToken(vt_test.provider, "foobar"))
}

func TestValidateTokenRequestNetworkFailure(t *testing.T) {
	vt_test := NewValidateTokenTest()
	// Close immediately to simulate a network failure
	vt_test.Close()
	assert.Equal(t, false, validateToken(vt_test.provider, "foobar"))
}

func TestValidateTokenExpiredToken(t *testing.T) {
	vt_test := NewValidateTokenTest()
	defer vt_test.Close()
	vt_test.response_code = 401
	assert.Equal(t, false, validateToken(vt_test.provider, "foobar"))
}
