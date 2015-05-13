package providers

import (
	"github.com/bitly/google_auth_proxy/api"
	"log"
)

func validateToken(p Provider, access_token string) bool {
	if access_token == "" || p.Data().ValidateUrl == nil {
		return false
	}
	url := p.Data().ValidateUrl.String() + "?access_token=" + access_token
	if resp, err := api.RequestUnparsedResponse(url, nil); err != nil {
		log.Printf("token validation request failed: %s", err)
		return false
	} else {
		return resp.StatusCode == 200
	}
}
