package providers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type redeemResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	IdToken      string `json:"id_token"`
}

func newRedeemServer(body []byte) (*url.URL, *httptest.Server, map[string]interface{}) {
	req := make(map[string]interface{})
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if req != nil {
			buf, _ := ioutil.ReadAll(r.Body)
			rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

			r.Body = rdr // OK since rdr2 implements the io.ReadCloser interface

			req["body"] = buf
			req["header"] = r.Header
			username, password, _ := r.BasicAuth()
			req["username"] = username
			req["password"] = password
		}
		rw.Write(body)
	}))
	u, _ := url.Parse(s.URL)
	return u, s, req
}
