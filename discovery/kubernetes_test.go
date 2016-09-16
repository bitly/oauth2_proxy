package discovery

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
)

type fakeHTTPHandler struct {
}

func (h *fakeHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
}

func fakeProxyAllocator(u *url.URL) http.Handler {
	return &fakeHTTPHandler{}
}

func newKubernetesTest(wg *sync.WaitGroup) (*k8sDiscovery, *watch.FakeWatcher) {
	k8s := &k8sDiscovery{
		proxyAllocator: fakeProxyAllocator,
		pathHandlers:   make(map[string]http.Handler),
		services:       make(map[string]*svcEndpoint),
	}

	watcher := watch.NewFake()
	wg.Add(1)

	go func() {
		k8s.run(watcher)
		wg.Done()
	}()

	return k8s, watcher
}

func TestServiceAdd(t *testing.T) {
	var wg sync.WaitGroup
	k8s, watcher := newKubernetesTest(&wg)

	watcher.Add(
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{Namespace: "default", Name: "foo"},
		})

	watcher.Add(
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Namespace:   "default",
				Name:        "bar",
				Annotations: map[string]string{OAuth2ProxyPathAnnotation: "xxx"},
			},
		})

	watcher.Stop()
	wg.Wait()

	if len(k8s.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(k8s.services))
	}
}

func TestServiceDelete(t *testing.T) {
	var wg sync.WaitGroup
	k8s, watcher := newKubernetesTest(&wg)

	services := []*v1.Service{
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Namespace:   "default",
				Name:        "foo",
				Annotations: map[string]string{OAuth2ProxyPathAnnotation: "xxx"},
			},
		},
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Namespace:   "default",
				Name:        "bar",
				Annotations: map[string]string{OAuth2ProxyPathAnnotation: "xxy"},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{v1.ServicePort{Port: 8080}},
			},
		},
	}

	for _, obj := range services {
		watcher.Add(obj)
	}

	watcher.Delete(
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{Namespace: "default", Name: "foo"},
		})

	watcher.Stop()
	wg.Wait()

	if len(k8s.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(k8s.services))
	}

	if endpoint, exists := k8s.services["default/bar"]; exists {
		if endpoint.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", endpoint.Port)
		}
	} else {
		t.Error("Service not found")
	}
}

func TestServiceChange(t *testing.T) {
	var wg sync.WaitGroup
	k8s, watcher := newKubernetesTest(&wg)

	svc := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Namespace:   "default",
			Name:        "foo",
			Annotations: map[string]string{OAuth2ProxyPathAnnotation: "xxx"},
		},
	}

	watcher.Add(svc)
	watcher.Stop()
	wg.Wait()

	watcher = watch.NewFake()
	wg.Add(1)
	go func() {
		k8s.run(watcher)
		wg.Done()
	}()

	svc.ObjectMeta.Annotations[OAuth2ProxyPathAnnotation] = "yyy"
	watcher.Modify(svc)

	watcher.Stop()
	wg.Wait()

	if len(k8s.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(k8s.services))
	}
	if len(k8s.pathHandlers) != 1 {
		t.Errorf("Expected 1 paths, got %d", len(k8s.pathHandlers))
	}
	path := svc.ObjectMeta.Annotations[OAuth2ProxyPathAnnotation]
	if _, exists := k8s.pathHandlers[path]; !exists {
		t.Error(path)
	}
}

func TestStatusTemplate(t *testing.T) {
	tmpl, err := template.New("test").Parse(statusTemplate)
	if err != nil {
		t.Fatal(err)
	}

	testData := map[string]*svcEndpoint{
		"default/foo": &svcEndpoint{"/foo/", -1},
		"default/bar": &svcEndpoint{"/bar/", 9090},
	}

	b := new(bytes.Buffer)
	err = tmpl.Execute(b, testData)
	if err != nil {
		t.Fatal(err)
	}
}
