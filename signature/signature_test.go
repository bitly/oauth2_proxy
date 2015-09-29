package signature

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
)

func newTestRequest(request ...string) (req *http.Request) {
	reqBuf := bufio.NewReader(
		strings.NewReader(strings.Join(request, "\n")))
	if req, err := http.ReadRequest(reqBuf); err != nil {
		panic(err)
	} else {
		return req
	}
}

func TestRequestSignaturePost(t *testing.T) {
	body := `{ "hello": "world!" }`
	req := newTestRequest(
		"POST /foo/bar HTTP/1.1",
		"Content-Length: "+strconv.Itoa(len(body)),
		"Content-MD5: deadbeef",
		"Content-Type: application/json",
		"Date: 2015-09-28",
		"Authorization: trust me",
		"X-Forwarded-User: mbland",
		"X-Forwarded-Email: mbland@acm.org",
		"X-Forwarded-Access-Token: feedbead",
		"Cookie: foo; bar; baz=quux",
		"Gap-Auth: mbland",
		"",
		body,
	)
	assert.Equal(t, StringToSign(req), strings.Join([]string{
		"POST",
		strconv.Itoa(len(body)),
		"deadbeef",
		"application/json",
		"2015-09-28",
		"trust me",
		"mbland",
		"mbland@acm.org",
		"feedbead",
		"foo; bar; baz=quux",
		"mbland",
		"/foo/bar",
	}, "\n"))
	assert.Equal(t, RequestSignature(req, "foobar"),
		"sha1 722UbRYfC6MnjtIxqEJMDPrW2mk=")
}

func TestRequestSignatureGet(t *testing.T) {
	req := newTestRequest(
		"GET /foo/bar HTTP/1.1",
		"Date: 2015-09-29",
		"Cookie: foo; bar; baz=quux",
		"Gap-Auth: mbland",
		"",
		"",
	)
	assert.Equal(t, StringToSign(req), strings.Join([]string{
		"GET",
		"",
		"",
		"",
		"2015-09-29",
		"",
		"",
		"",
		"",
		"foo; bar; baz=quux",
		"mbland",
		"/foo/bar",
	}, "\n"))
	assert.Equal(t, RequestSignature(req, "foobar"),
		"sha1 JBQJcmSTteQyHZXFUA9glis9BIk=")
}

func newGetRequest() *http.Request {
	return newTestRequest(
		"GET /foo/bar HTTP/1.1",
		"Date: 2015-09-29",
		"Cookie: foo; bar; baz=quux",
		"Gap-Auth: mbland",
		"",
		"",
	)
}

func TestValidateRequestNoSignature(t *testing.T) {
	req := newGetRequest()
	result, header, computed := ValidateRequest(req, "foobar")
	assert.Equal(t, result, NO_SIGNATURE)
	assert.Equal(t, header, "")
	assert.Equal(t, computed, "")
}

func TestValidateRequestInvalidFormat(t *testing.T) {
	req := newGetRequest()
	badValue := "should be algorithm and digest value"
	req.Header.Set("GAP-Signature", badValue)
	result, header, computed := ValidateRequest(req, "foobar")
	assert.Equal(t, result, INVALID_FORMAT)
	assert.Equal(t, header, badValue)
	assert.Equal(t, computed, "")
}

func TestValidateRequestUnsupportedAlgorithm(t *testing.T) {
	req := newGetRequest()
	validSignature := RequestSignature(req, "foobar")
	components := strings.Split(validSignature, " ")
	signatureWithUnsupportedAlgorithm := "unsupported " + components[1]
	req.Header.Set("GAP-Signature", signatureWithUnsupportedAlgorithm)
	result, header, computed := ValidateRequest(req, "foobar")
	assert.Equal(t, result, UNSUPPORTED_ALGORITHM)
	assert.Equal(t, header, signatureWithUnsupportedAlgorithm)
	assert.Equal(t, computed, "")
}

func TestValidateRequestMatch(t *testing.T) {
	req := newGetRequest()
	expectedSignature := RequestSignature(req, "foobar")
	req.Header.Set("GAP-Signature", expectedSignature)
	result, header, computed := ValidateRequest(req, "foobar")
	assert.Equal(t, result, MATCH)
	assert.Equal(t, header, expectedSignature)
	assert.Equal(t, computed, expectedSignature)
}

func TestValidateRequestMismatch(t *testing.T) {
	req := newGetRequest()
	foobarSignature := RequestSignature(req, "foobar")
	barbazSignature := RequestSignature(req, "barbaz")
	req.Header.Set("GAP-Signature", foobarSignature)
	result, header, computed := ValidateRequest(req, "barbaz")
	assert.Equal(t, result, MISMATCH)
	assert.Equal(t, header, foobarSignature)
	assert.Equal(t, computed, barbazSignature)
}
