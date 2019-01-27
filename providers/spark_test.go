package providers

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/bmizerany/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func testSparkProvider(hostname string) *SparkProvider {

	p := NewSparkProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}
	return p
}

func TestSparkProviderDefaults(t *testing.T) {
	p := testSparkProvider("")
	assert.NotEqual(t, nil, p)
	assert.Equal(t, reflect.TypeOf(p.SpaceID).Kind(), reflect.String)
	p.SetSparkSpaceID("1234")
	assert.Equal(t, p.SpaceID, "1234")
	assert.Equal(t, "CiscoSpark", p.Data().ProviderName)
	assert.Equal(t, "https://api.ciscospark.com/v1/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://api.ciscospark.com/v1/access_token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://api.ciscospark.com/v1/people/me",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://api.ciscospark.com/v1/tokeninfo",
		p.Data().ValidateURL.String())
}

func TestSparkProviderOverrides(t *testing.T) {
	p := NewSparkProvider(
		&ProviderData{
			LoginURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/v1/authorize"},
			RedeemURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/v1/access_token"},
			ProfileURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/v1/people/me"},
			ValidateURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/v1/tokeninfo"},
			Scope: "profile"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "CiscoSpark", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/v1/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/v1/access_token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/v1/people/me",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://example.com/v1/tokeninfo",
		p.Data().ValidateURL.String())
	assert.Equal(t, "profile", p.Data().Scope)
}

func TestSparkProviderGetEmailAddressNoSpaceID(t *testing.T) {

	session := &SessionState{AccessToken: "imaginary_good_access_token"}
	p := testSparkProvider("")

	_, err := p.GetEmailAddress(session)

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "SpaceID required", err.Error())
}

func TestSparkProviderGetEmailAddress(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://api.ciscospark.com").
		Get("/v1/rooms/abc123").
		Reply(200).
		BodyString(`{
	"id": "abc123",
	"title": "testing 123",
	"type": "group",
	"isLocked": false,
	"sipAddress": "157228592@foo.com",
	"lastActivity": "2018-02-06T17:07:17.256Z","creatorId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9kNWViYTg3MS1iZGM3LTQ5N2UtOTY3ZS05NzkyZjFmYmQzZTU","created": "2018-02-02T21:30:23.751Z"}`)

	gock.New("https://api.ciscospark.com").
		Get("/v1/people/me").
		Reply(200).
		BodyString(`{"id":"Y2lzY29zcGFyazovL3VzL1BFT1BMRS82ZmQzNjBkOC03NzIyLTQ2NTUtYThiNC05NGRkNTBkNTc5Nzc","emails":["ciscouser@ciscospark.com"],"displayName":"johndyer","avatar":"https://c74213ddaf67eb02dabb-04de5163e3f90393a9f7bb6f7f0967f1.ssl.cf1.rackcdn.com/V1~c2e6d34e79147974acfcd30a09aef889~qZcXkE0_TDq6pmcTn6CUJA==~1600","created":"2016-01-15T20:29:03.498Z"}`)

	session := &SessionState{AccessToken: "imaginary_good_access_token"}
	p := testSparkProvider("")

	p.SetSparkSpaceID("abc123")
	email, err := p.GetEmailAddress(session)

	assert.Equal(t, nil, err)
	assert.Equal(t, "ciscouser@ciscospark.com", email)
}

func TestSparkProviderBadData(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://api.ciscospark.com").
		Get("/v1/rooms/abc123").
		Reply(200).
		BodyString(`NOT_JSON`)

	session := &SessionState{AccessToken: "imaginary_good_access_token"}
	p := testSparkProvider("")

	p.SetSparkSpaceID("abc123")
	email, err := p.GetEmailAddress(session)

	assert.Equal(t, "invalid character 'N' looking for beginning of value", err.Error())
	assert.Equal(t, "", email)
}

func TestSparkProviderGetEmailAddressNoAccessToken(t *testing.T) {

	p := testSparkProvider("")

	session := &SessionState{}

	email, err := p.GetEmailAddress(session)

	assert.Equal(t, "SpaceID required", err.Error())
	assert.Equal(t, "", email)
}

func TestSparkGetTeamIDLookupRequestFailedRequest(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.ciscospark.com").
		Get("/v1/rooms/abc123").
		Reply(401).
		BodyString(`error`)

	gock.New("https://api.ciscospark.com").
		Get("/v1/people/me").
		Reply(401).
		BodyString(`error`)

	p := testSparkProvider("")
	p.SetSparkSpaceID("abc123")
	session := &SessionState{}

	// We'll trigger a request failure by using an unexpected access
	// token. Alternatively, we could allow the parsing of the payload as
	// JSON to fail.
	email, err := p.GetEmailAddress(session)

	assert.Equal(t, "got 401 error", err.Error())
	assert.Equal(t, "", email)
}
