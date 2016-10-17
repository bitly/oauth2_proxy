package backends

import (
	"log"
	"net/http"
	"net/url"

	"github.com/smartystreets/go-aws-auth"
)

func registerNewAwsBackend(u *url.URL, opts *Options, serveMux *http.ServeMux) {
	path := u.Path
	u.Path = ""
	log.Printf("mapping path %q => aws-upstream %q", path, u)

	proxy := NewReverseProxy(u)
	serveMux.Handle(path, &awsProxy{u, proxy, opts})
}

type awsProxy struct {
	upstream *url.URL
	handler  http.Handler
	options  *Options
}

func (a *awsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = a.upstream.Host
	r.URL.Scheme = a.upstream.Scheme
	r.Host = a.upstream.Host

	awsauth.Sign(r, awsauth.Credentials{
		AccessKeyID:     a.options.AwsAccessKeyId,
		SecretAccessKey: a.options.AwsSecretAccessKey,
	})
	a.handler.ServeHTTP(w, r)
}
