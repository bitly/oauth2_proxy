package providers

import (
	"bytes"
	"encoding/json"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type WildApricotProvider struct {
	*ProviderData
	RedeemRefreshURL *url.URL
	// GroupValidator is a function that determines if the passed email is in
	// the configured WildApricot group.
	GroupValidator func(string) bool
}

func NewWildApricotProvider(p *ProviderData) *WildApricotProvider {
	if p.Scope == "" {
		p.Scope = "auto"
	}

	return &WildApricotProvider{
		ProviderData: p,
		// Set a default GroupValidator to just always return valid (true), it will
		// be overwritten if we configured a WildApricot group restriction.
		GroupValidator: func(email string) bool {
			return true
		},
	}
}

func (p *WildApricotProvider) Redeem(redirectURL, code string) (s *SessionState, err error) {
	if code == "" {
		err = errors.New("missing code")
		return
	}

	params := url.Values{}
	params.Add("redirect_uri", redirectURL)
	params.Add("client_id", p.ClientID)
	params.Add("scope", p.Scope)
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")
	var req *http.Request
	req, err = http.NewRequest("POST", p.RedeemURL.String(), bytes.NewBufferString(params.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization","Basic " + base64.StdEncoding.EncodeToString([]byte(p.ClientID +":"+p.ClientSecret)))

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
		Permissions []struct {
			AccountID int64 `json:"AccountID"`
		} `json:"Permissions"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return
	}

	req, err = http.NewRequest("GET", "https://api.wildapricot.org/v2/Accounts/" + fmt.Sprintf("%v", jsonResponse.Permissions[0].AccountID) + "/Contacts/me", nil)
	req.Header.Add("Authorization","Bearer " + jsonResponse.AccessToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("got %d from %q %s", resp.StatusCode, p.RedeemURL.String(), body)
		return
	}

	var contactResponse struct {
		FirstName string `json:"FirstName"`
		LastName string `json:"LastName"`
		DisplayName string `json:"DisplayName"`
		Email string `json:"Email"`
		MembershipLevel struct {
			Name string `json:"Name"`
		} `json:"MembershipLevel"`
	}
	err = json.Unmarshal(body, &contactResponse)
	if err != nil {
		return
	}


	s = &SessionState{
		AccessToken:  jsonResponse.AccessToken,
		ExpiresOn:    time.Now().Add(time.Duration(jsonResponse.ExpiresIn) * time.Second).Truncate(time.Second),
		RefreshToken: jsonResponse.RefreshToken,
		Email:        contactResponse.Email,
	}
	return
}
