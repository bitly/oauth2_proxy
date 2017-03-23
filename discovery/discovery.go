package discovery

import (
	"net/http"
	"net/url"
)

// Discovery defines the interface for backend discovery
type Discovery interface {
	NewServeMux(http.Handler) http.Handler
}

type nullDiscovery struct {
}

func (n *nullDiscovery) NewServeMux(mux http.Handler) http.Handler {
	return mux
}

// Create allocates the discovery provider based on the configuration.
func Create(provider string, proxyAllocator func(u *url.URL) http.Handler) Discovery {
	switch provider {
	case "kubernetes":
		return newKubernetesDiscovery(proxyAllocator)
	default:
		return &nullDiscovery{}
	}
}
