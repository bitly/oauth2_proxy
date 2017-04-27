package providers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

type BitBucketProvider struct {
	*ProviderData
}

func NewBitBucketProvider(p *ProviderData) *BitBucketProvider {
	p.ProviderName = "BitBucket"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "bitbucket.org",
			Path:   "/site/oauth2/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "bitbucket.org",
			Path:   "/site/oauth2/access_token",
		}
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "api.bitbucket.org",
			Path:   "/2.0/user/emails",
		}
	}
	if p.Scope == "" {
		p.Scope = "account team"
	}
	return &BitBucketProvider{ProviderData: p}
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}

func (p *BitBucketProvider) GetEmailAddress(s *SessionState) (string, error) {

	var emails struct {
		Values []struct {
			Email   string `json:"email"`
			Primary bool   `json:"is_primary"`
		}
	}
	req, err := http.NewRequest("GET",
		p.ValidateURL.String()+"?access_token="+s.AccessToken, nil)
	if err != nil {
		log.Printf("failed building request %s", err)
		return "", err
	}
	err = api.RequestJson(req, &emails)
	if err != nil {
		log.Printf("failed making request %s", err)
		debug(httputil.DumpRequestOut(req, true))
		return "", err
	}

	for _, email := range emails.Values {
		if email.Primary {
			log.Printf("got here, returning %s\n", email.Email)
			return email.Email, nil
		}
	}

	return "", nil
}
