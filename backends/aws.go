package backends

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/smartystreets/go-aws-auth"
)

func registerNewAwsBackend(u *url.URL, opts *Options, serveMux *http.ServeMux) {
	path := u.Path
	u.Path = ""
	log.Printf("mapping path %q => upstream %q", path, u)

	proxy := httputil.NewSingleHostReverseProxy(u)
	serveMux.Handle(path, &awsProxy{u, proxy})
}

type awsProxy struct {
	upstream *url.URL
	handler  http.Handler
}

func (a *awsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = a.upstream.Host
	r.URL.Scheme = a.upstream.Scheme
	r.Host = a.upstream.Host
	awsauth.Sign4(r)
	a.handler.ServeHTTP(w, r)
}
