package backends

import (
	"crypto"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/18F/hmacauth"
)

const (
	BackendTypeAws     = "aws"
	BackendTypeDefault = "default"
)

func Register(backendType string, u *url.URL, opts *Options, serveMux *http.ServeMux) {
	switch backendType {
	case BackendTypeDefault:
		registerNewDefaultBackend(u, opts, serveMux)
	case BackendTypeAws:
		registerNewAwsBackend(u, opts, serveMux)
	default:
		panic(fmt.Errorf("Invalid backendType: %s", backendType))
	}
}

// Default Backend Methods

// GAP-Auth Signatures
const GAPSignatureHeader = "GAP-Signature"

var GAPSignatureHeaders []string = []string{
	"Content-Length",
	"Content-Md5",
	"Content-Type",
	"Date",
	"Authorization",
	"X-Forwarded-User",
	"X-Forwarded-Email",
	"X-Forwarded-Access-Token",
	"Cookie",
	"Gap-Auth",
}

type GAPSignatureData struct {
	Hash crypto.Hash
	Key  string
}

type Options struct {
	SignatureData      *GAPSignatureData
	PassHostHeader     bool
	AwsAccessKeyId     string
	AwsSecretAccessKey string
}

// this is what is used to handle urls in the "upstreams" option flag
func registerNewDefaultBackend(u *url.URL, opts *Options, serveMux *http.ServeMux) {
	// handle gap auth
	var auth hmacauth.HmacAuth
	if sigData := opts.SignatureData; sigData != nil {
		auth = hmacauth.NewHmacAuth(sigData.Hash, []byte(sigData.Key),
			GAPSignatureHeader, GAPSignatureHeaders)
	}
	path := u.Path
	switch u.Scheme {
	case "http", "https":
		u.Path = ""
		log.Printf("mapping path %q => upstream %q", path, u)
		proxy := NewReverseProxy(u)
		if !opts.PassHostHeader {
			setProxyUpstreamHostHeader(proxy, u)
		} else {
			setProxyDirector(proxy)
		}
		serveMux.Handle(path,
			&UpstreamProxy{u, proxy, auth})
	case "file":
		if u.Fragment != "" {
			path = u.Fragment
		}
		log.Printf("mapping path %q => file system %q", path, u.Path)
		proxy := NewFileServer(path, u.Path)
		serveMux.Handle(path, &UpstreamProxy{u, proxy, nil})
	default:
		panic(fmt.Sprintf("unknown upstream protocol %s", u.Scheme))
	}
}

type UpstreamProxy struct {
	upstream *url.URL
	handler  http.Handler
	auth     hmacauth.HmacAuth
}

func (u *UpstreamProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("GAP-Upstream-Address", u.upstream.Host)
	if u.auth != nil {
		r.Header.Set("GAP-Auth", w.Header().Get("GAP-Auth"))
		u.auth.SignRequest(r)
	}
	u.handler.ServeHTTP(w, r)
}

func NewReverseProxy(target *url.URL) (proxy *httputil.ReverseProxy) {
	return httputil.NewSingleHostReverseProxy(target)
}

func setProxyUpstreamHostHeader(proxy *httputil.ReverseProxy, target *url.URL) {
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		// use RequestURI so that we aren't unescaping encoded slashes in the request path
		req.Host = target.Host
		req.URL.Opaque = req.RequestURI
		req.URL.RawQuery = ""
	}
}

func setProxyDirector(proxy *httputil.ReverseProxy) {
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		// use RequestURI so that we aren't unescaping encoded slashes in the request path
		req.URL.Opaque = req.RequestURI
		req.URL.RawQuery = ""
	}
}

func NewFileServer(path string, filesystemPath string) (proxy http.Handler) {
	return http.StripPrefix(path, http.FileServer(http.Dir(filesystemPath)))
}
