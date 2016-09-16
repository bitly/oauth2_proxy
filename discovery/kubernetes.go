package discovery

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
	"k8s.io/client-go/1.4/rest"
)

type svcEndpoint struct {
	Path string
	Port int32
}

type k8sDiscovery struct {
	proxyAllocator func(u *url.URL) http.Handler
	pathHandlers   map[string]http.Handler
	services       map[string]*svcEndpoint
	defaultHandler http.Handler
	statusTmpl     *template.Template
}

const (
	// OAuth2ProxyPathAnnotation is the annotation used by a kubernetes service
	// to add a specific string to the list of paths handled by this proxy.
	OAuth2ProxyPathAnnotation = "oauth2-proxy/path"

	// OAuth2ProxyPortAnnotation specifies the HTTP port to forward traffic to.
	OAuth2ProxyPortAnnotation = "oauth2-proxy/port"

	discoveryStatusPage = "/oauth2/discovery"
)

// ServeHttp implements the http.Handler interface.
// It is called to demux request paths.
func (k *k8sDiscovery) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case discoveryStatusPage:
		k.statusPage(rw, req)
		return
	}

	var bestMatch string
	handler := k.defaultHandler
	for k, v := range k.pathHandlers {
		if strings.HasPrefix(req.URL.Path, k) && len(k) > len(bestMatch) {
			bestMatch = k
			handler = v
		}
	}

	handler.ServeHTTP(rw, req)
}

func (k *k8sDiscovery) NewServeMux(mux http.Handler) http.Handler {
	k.defaultHandler = mux
	return k
}

const statusTemplate = `<!DOCTYPE html>
<html lang="en" charset="utf-8">
<head>
	<title>Kubernetes Service Discovery</title>
	<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
	<style>
	body {
		font-family: "Helvetica Neue",Helvetica,Arial,sans-serif;
		font-size: 14px;
		line-height: 1.42857143;
		color: #333;
		background: #f0f0f0;
	}
    div.centered {
        text-align: center;
    }
    div.centered table {
        margin: 0 auto;
        text-align: left;
    }
    </style>
</head>
<body>
<div class="centered">
<table>
    <thead>
        <tr>
            <th>Service</th>
            <th>Port</th>
            <th>Path</th>
        </tr>
    </thead>
    <tbody>
{{ range $key, $value := . }}
    <tr>
        <td>{{ $key }}</td>
        <td>{{ if ge $value.Port 0 }}{{ $value.Port }}{{ end }}</td>
        <td><a href="{{ $value.Path }}">{{ $value.Path }}</a></td>
    </tr>
{{ end }}
    </tbody>
</table>
</div>
</body>
</html>
`

func (k *k8sDiscovery) statusPage(w http.ResponseWriter, r *http.Request) {
	k.statusTmpl.Execute(w, k.services)
}

func getServicePort(svc *v1.Service) (int32, bool) {
	if port, exists := svc.Annotations[OAuth2ProxyPortAnnotation]; exists {
		p, err := strconv.Atoi(port)
		if err != nil {
			log.Printf("Invalid annotation (%s) for %s/%s", port, svc.Namespace, svc.Name)
			return -1, false
		}
		return int32(p), true
	} else if len(svc.Spec.Ports) == 1 {
		svcPort := svc.Spec.Ports[0]
		if svcPort.Port != 80 {
			return svcPort.Port, true
		}
	}
	return -1, false
}

func makeSvcEndpoint(svc *v1.Service) *svcEndpoint {
	path, exists := svc.Annotations[OAuth2ProxyPathAnnotation]
	if !exists {
		return nil
	}
	endpoint := &svcEndpoint{
		Path: path,
		Port: -1,
	}
	if port, isSet := getServicePort(svc); isSet {
		endpoint.Port = port
	}
	return endpoint
}

func makeServiceURL(svc *v1.Service, endpoint *svcEndpoint) *url.URL {
	schemeHost := fmt.Sprintf("http://%s.%s.svc", svc.Name, svc.Namespace)
	if endpoint.Port >= 0 {
		schemeHost += fmt.Sprintf(":%d", endpoint.Port)
	}

	u, err := url.Parse(schemeHost)
	if err != nil {
		log.Println(err)
		return nil
	}
	return u
}

func (k *k8sDiscovery) serviceAdd(svc *v1.Service) {
	endpoint := makeSvcEndpoint(svc)
	if endpoint == nil {
		return
	}

	svcID := svc.Namespace + "/" + svc.Name
	log.Println("ADD service", svcID)

	if prev, dup := k.services[svcID]; dup {
		log.Printf("ADD event for existing service %s", svcID)
		if reflect.DeepEqual(endpoint, prev) {
			return
		}
		delete(k.pathHandlers, endpoint.Path)
	}

	if _, dup := k.pathHandlers[endpoint.Path]; dup {
		log.Printf("Duplicate %s annotation for %s: %s/%s", OAuth2ProxyPathAnnotation, endpoint.Path, svc.Namespace, svc.Name)
		return
	}

	u := makeServiceURL(svc, endpoint)
	k.pathHandlers[endpoint.Path] = k.proxyAllocator(u)
	k.services[svcID] = endpoint
}

func (k *k8sDiscovery) serviceDelete(svc *v1.Service) {
	svcID := svc.Namespace + "/" + svc.Name
	log.Println("DELETE service", svcID)
	if endpoint, exists := k.services[svcID]; exists {
		delete(k.services, svcID)
		delete(k.pathHandlers, endpoint.Path)
	}
}

func (k *k8sDiscovery) serviceChange(svc *v1.Service) {
	svcID := svc.Namespace + "/" + svc.Name
	prev := k.services[svcID]
	endpoint := makeSvcEndpoint(svc)

	if prev != nil && endpoint != nil {
		if reflect.DeepEqual(prev, endpoint) {
			return
		}

		log.Println("CHANGE service", svcID)
		u := makeServiceURL(svc, endpoint)
		k.pathHandlers[endpoint.Path] = k.proxyAllocator(u)
		delete(k.pathHandlers, prev.Path)
		k.services[svcID] = endpoint
	} else if endpoint != nil {
		k.serviceAdd(svc)
	} else if prev != nil {
		k.serviceDelete(svc)
	}
}

func (k *k8sDiscovery) run(watcher watch.Interface) {
	var shutdown bool
	for !shutdown {
		select {
		case ev, ok := <-watcher.ResultChan():
			if !ok {
				shutdown = true
				break
			}
			switch ev.Type {
			case watch.Added:
				k.serviceAdd(ev.Object.(*v1.Service))
			case watch.Deleted:
				k.serviceDelete(ev.Object.(*v1.Service))
			case watch.Modified:
				k.serviceChange(ev.Object.(*v1.Service))
			}
		}
	}
}

func newKubernetesDiscovery(proxyAllocator func(u *url.URL) http.Handler) *k8sDiscovery {
	log.Println("Enabling kubernetes service discovery")

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := clientset.Core().Services("").Watch(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("status").Parse(statusTemplate)
	if err != nil {
		log.Fatal(err)
	}
	k8s := &k8sDiscovery{
		proxyAllocator: proxyAllocator,
		pathHandlers:   make(map[string]http.Handler),
		services:       make(map[string]*svcEndpoint),
		statusTmpl:     tmpl,
	}

	go k8s.run(watcher)

	return k8s
}
