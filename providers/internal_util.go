package providers

import (
	"github.com/bitly/google_auth_proxy/api"
	"log"
)

func validateToken(p Provider, access_token string,
	headers map[string]string) bool {
	if access_token == "" || p.Data().ValidateUrl == nil {
		return false
	}
	url := p.Data().ValidateUrl.String()
	if len(headers) == 0 {
		url = url + "?access_token=" + access_token
	}
	if resp, err := api.RequestUnparsedResponse(url, headers); err != nil {
		log.Printf("token validation request failed: %s", err)
		return false
	} else {
		return resp.StatusCode == 200
	}
}
