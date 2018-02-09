package providers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

// SparkProvider ...  Top level provider
type SparkProvider struct {
	*ProviderData
	SpaceID string
}

// NewSparkProvider - Instantiate provider
func NewSparkProvider(p *ProviderData) *SparkProvider {

	p.ProviderName = "CiscoSpark"
	if p.LoginURL.String() == "" {

		p.LoginURL = &url.URL{Scheme: "https",
			Host: "api.ciscospark.com",
			Path: "/v1/authorize"}
	}

	if p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{Scheme: "https",
			Host: "api.ciscospark.com",
			Path: "/v1/access_token"}
	}
	if p.ProfileURL.String() == "" {
		p.ProfileURL = &url.URL{Scheme: "https",
			Host: "api.ciscospark.com",
			Path: "/v1/people/me"}
	}
	if p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{Scheme: "https",
			Host: "api.ciscospark.com",
			Path: "/v1/tokeninfo"}
	}
	if p.Scope == "" {
		p.Scope = "spark:people_read"
	}
	return &SparkProvider{ProviderData: p}
}

// SetSparkSpaceID - Set the SpaceID to use as a filter
func (p *SparkProvider) SetSparkSpaceID(SpaceID string) {
	p.SpaceID = SpaceID
}

// GetEmailAddress - Uses Oauth API to fetch users profile, including EmailAddress
func (p *SparkProvider) GetEmailAddress(s *SessionState) (string, error) {
	if p.SpaceID == "" {
		return "", fmt.Errorf("SpaceID required")
	}

	log.Printf("### Cisco Spark ### -  Check if user belongs to space\n")

	roomEndpoint := &url.URL{
		Scheme: p.ValidateURL.Scheme,
		Host:   p.ValidateURL.Host,
		Path:   "/v1/rooms/" + p.SpaceID,
	}
	req, _ := http.NewRequest("GET", roomEndpoint.String(), nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))

	j, err := api.Request(req)
	if err != nil {
		log.Printf("### Cisco Spark ### - failed making roomEndpoint request %s\n", err)
		return "", err
	}

	spaceID, err := j.Get("id").String()
	if err != nil {
		log.Printf("### Cisco Spark ### - Failed to get id %s", err)
		return "", err
	}

	log.Printf("SpaceID [%s] found\n", spaceID)

	peopleEndpoint := &url.URL{
		Scheme: p.ValidateURL.Scheme,
		Host:   p.ValidateURL.Host,
		Path:   "/v1/people/me",
	}

	req, _ = http.NewRequest("GET", peopleEndpoint.String(), nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))

	json, err := api.Request(req)
	if err != nil {
		log.Printf("### Cisco Spark ### failed making peopleEndpoint request %s\n", err)
		return "", err
	}

	email, err := json.Get("emails").GetIndex(0).String()
	if err != nil {
		return "", err
	}
	log.Printf("### Cisco Spark ### -  email %s\n", email)
	return email, nil
}
