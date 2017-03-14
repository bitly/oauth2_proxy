package providers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
	"github.com/bitly/go-simplejson"
	"fmt"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"time"
	"errors"
)

type JhipsterUaaProvider struct {
	*ProviderData
	Authority string
}

func NewJhipsterUaaProvider(p *ProviderData) *JhipsterUaaProvider {
	p.ProviderName = "JhipsterUaa"

	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/oauth/authorize",
		}
	}

	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/oauth/token",
		}
	}

	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/api/account",
		}
	}

	if p.Scope == "" {
		p.Scope = "openid"
	}

	return &JhipsterUaaProvider{ProviderData: p}
}

func getJhipsterUaaHeader(access_token string) http.Header {
	header := make(http.Header)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	header.Set("Accept", "application/json")
	return header
}

func (p *JhipsterUaaProvider) SetAuthority(authority string) {
	p.Authority = authority
}

func (p *JhipsterUaaProvider) Redeem(redirectURL, code string) (s *SessionState, err error) {
	if code == "" {
		err = errors.New("missing code")
		return
	}

	params := url.Values{}
	params.Add("redirect_uri", redirectURL)
	params.Add("client_id", p.ClientID)
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")
	var req *http.Request
	req, err = http.NewRequest("POST", p.RedeemURL.String(), bytes.NewBufferString(params.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("got %d from %q %s", resp.StatusCode, p.RedeemURL.String(), body)
		return
	}

	var jsonResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`

	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return
	}

	s = &SessionState{
		AccessToken:  jsonResponse.AccessToken,
		ExpiresOn:    time.Now().Add(time.Duration(jsonResponse.ExpiresIn) * time.Second).Truncate(time.Second),
		RefreshToken: jsonResponse.RefreshToken,
	}
	return
}

func (p *JhipsterUaaProvider) hasAuthority(jsonResponse *simplejson.Json) (bool, error) {

	authorities, err := jsonResponse.Get("authorities").Array()
	if err != nil {
		log.Printf("failed getting authorities array: %s", err)
		return false, err
	}

	if len(authorities) > 0 {
		for i := range authorities {
			if(authorities[i] == p.Authority) {
				return true, nil
			}
		}
	}

	log.Printf("Authority %s not found in %s", p.Authority, authorities)
	return false, nil
}


func (p *JhipsterUaaProvider) GetEmailAddress(s *SessionState) (string, error) {

	req, err := http.NewRequest("GET", p.ValidateURL.String(), nil)
	if err != nil {
		log.Printf("failed building request %s", err)
		return "", err
	}

	req.Header = getJhipsterUaaHeader(s.AccessToken)

	jsonResponse, err := api.Request(req)
	if err != nil {
		log.Printf("failed making request %s", err)
		return "", err
	}

	// if we require an authority, check that first
	if p.Authority != "" {
		if ok, err := p.hasAuthority(jsonResponse); err != nil || !ok {
			return "", err
		}
	}

	return jsonResponse.Get("email").String()
}

func (p *JhipsterUaaProvider) RefreshSessionIfNeeded(s *SessionState) (bool, error) {
	if s == nil || s.ExpiresOn.After(time.Now()) || s.RefreshToken == "" {
		return false, nil
	}

	newToken, duration, err := p.redeemRefreshToken(s.RefreshToken)
	if err != nil {
		return false, err
	}

	// re-check that the user is in the proper google group(s)
	if !p.ValidateGroup(s.Email) {
		return false, fmt.Errorf("%s is no longer in the group(s)", s.Email)
	}

	origExpiration := s.ExpiresOn
	s.AccessToken = newToken
	s.ExpiresOn = time.Now().Add(duration).Truncate(time.Second)
	log.Printf("refreshed access token %s (expired on %s)", s, origExpiration)
	return true, nil
}

func (p *JhipsterUaaProvider) redeemRefreshToken(refreshToken string) (token string, expires time.Duration, err error) {
	// https://developers.google.com/identity/protocols/OAuth2WebServer#refresh
	params := url.Values{}
	params.Add("client_id", p.ClientID)
	params.Add("client_secret", p.ClientSecret)
	params.Add("refresh_token", refreshToken)
	params.Add("grant_type", "refresh_token")
	var req *http.Request
	req, err = http.NewRequest("POST", p.RedeemURL.String(), bytes.NewBufferString(params.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("got %d from %q %s", resp.StatusCode, p.RedeemURL.String(), body)
		return
	}

	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}
	token = data.AccessToken
	expires = time.Duration(data.ExpiresIn) * time.Second
	return
}