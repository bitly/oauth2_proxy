package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

type OpenshiftProvider struct {
	*ProviderData
	Project string
}

func NewOpenshiftProvider(p *ProviderData) *OpenshiftProvider {
	p.ProviderName = "Openshift"
	if p.Scope == "" {
		p.Scope = "user:info user:check-access"
	}
	return &OpenshiftProvider{ProviderData: p}
}
func (p *OpenshiftProvider) SetProject(project string) {
	p.Project = project
	if project != "" {
		p.Scope += " user:list-projects"
	}
}

func (p *OpenshiftProvider) hasProject(accessToken string) (bool, error) {
	var projects struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
		Metadata   struct {
			SelfLink string `json:"selfLink"`
		} `json:"metadata"`
		Items []struct {
			Metadata struct {
				Name              string    `json:"name"`
				SelfLink          string    `json:"selfLink"`
				UID               string    `json:"uid"`
				ResourceVersion   string    `json:"resourceVersion"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
				Annotations       struct {
					OpenshiftIoDescription             string `json:"openshift.io/description"`
					OpenshiftIoDisplayName             string `json:"openshift.io/display-name"`
					OpenshiftIoRequester               string `json:"openshift.io/requester"`
					OpenshiftIoSaSccMcs                string `json:"openshift.io/sa.scc.mcs"`
					OpenshiftIoSaSccSupplementalGroups string `json:"openshift.io/sa.scc.supplemental-groups"`
					OpenshiftIoSaSccUIDRange           string `json:"openshift.io/sa.scc.uid-range"`
				} `json:"annotations"`
			} `json:"metadata"`
			Spec struct {
				Finalizers []string `json:"finalizers"`
			} `json:"spec"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	endpoint := &url.URL{
		Scheme: p.ValidateURL.Scheme,
		Host:   p.ValidateURL.Host,
		Path:   path.Join(p.ValidateURL.Path, "oapi/v1/projects/"),
	}
	req, _ := http.NewRequest("GET", endpoint.String(), nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, fmt.Errorf(
			"got %d from %q %s", resp.StatusCode, endpoint.String(), body)
	}

	if err := json.Unmarshal(body, &projects); err != nil {
		return false, fmt.Errorf("%s unmarshaling %s", err, body)
	}

	var hasProject bool
	for _, project := range projects.Items {
		if p.Project == project.Metadata.Name {
			hasProject = true
		}
	}
	if hasProject {
		log.Printf("User is part of project:%q", p.Project)
		return true, nil
	} else {
		log.Printf("User is not part of project:%q", p.Project)
	}
	return false, fmt.Errorf("User is not part of project:%q", p.Project)
}

func (p *OpenshiftProvider) GetEmailAddress(s *SessionState) (string, error) {

	var user struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
		Metadata   struct {
			Name              string    `json:"name"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		} `json:"metadata"`
		Identities []string      `json:"identities"`
		Groups     []interface{} `json:"groups"`
	}

	if p.Project != "" {
		if ok, err := p.hasProject(s.AccessToken); err != nil || !ok {
			return "", err
		}
	}

	endpoint := &url.URL{
		Scheme: p.ValidateURL.Scheme,
		Host:   p.ValidateURL.Host,
		Path:   path.Join(p.ValidateURL.Path, "oapi/v1/users/~"),
	}
	req, _ := http.NewRequest("GET", endpoint.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got %d from %q %s",
			resp.StatusCode, endpoint.String(), body)
	} else {
		log.Printf("got %d from %q %s", resp.StatusCode, endpoint.String(), body)
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return "", fmt.Errorf("%s unmarshaling %s", err, body)
	}
	if user.Metadata.Name != "" {
		return user.Metadata.Name, nil
	} else {
		return "", fmt.Errorf("unexpected response, no user present in %s from %q",
			body, endpoint.String())
	}
}
