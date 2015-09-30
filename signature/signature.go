package signature

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
)

var supportedAlgorithms map[string]crypto.Hash
var algorithmName map[crypto.Hash]string

func init() {
	supportedAlgorithms = map[string]crypto.Hash{
		"sha1": crypto.SHA1,
	}

	algorithmName = make(map[crypto.Hash]string)
	for name, algorithm := range supportedAlgorithms {
		algorithmName[algorithm] = name
	}
}

// The string to sign is based on the following request elements, inspired by:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html
func StringToSign(req *http.Request) string {
	return strings.Join([]string{
		req.Method,
		req.Header.Get("Content-Length"),
		req.Header.Get("Content-Md5"),
		req.Header.Get("Content-Type"),
		req.Header.Get("Date"),
		req.Header.Get("Authorization"),
		req.Header.Get("X-Forwarded-User"),
		req.Header.Get("X-Forwarded-Email"),
		req.Header.Get("X-Forwarded-Access-Token"),
		req.Header.Get("Cookie"),
		req.Header.Get("Gap-Auth"),
		req.URL.String(),
	}, "\n")
}

type unsupportedAlgorithm struct {
	algorithm string
}

func (e unsupportedAlgorithm) Error() string {
	return "unsupported request signature algorithm: " + e.algorithm
}

func HashAlgorithm(algorithm string) (result crypto.Hash, err error) {
	if result = supportedAlgorithms[algorithm]; result == crypto.Hash(0) {
		err = unsupportedAlgorithm{algorithm}
	}
	return
}

func RequestSignature(req *http.Request, hashAlgorithm crypto.Hash,
	secretKey string) string {
	h := hmac.New(hashAlgorithm.New, []byte(secretKey))
	h.Write([]byte(StringToSign(req)))

	if req.ContentLength != -1 && req.Body != nil {
		buf := make([]byte, req.ContentLength, req.ContentLength)
		req.Body.Read(buf)
		h.Write(buf)
	}

	var sig []byte
	sig = h.Sum(sig)
	return algorithmName[hashAlgorithm] + " " +
		base64.URLEncoding.EncodeToString(sig)
}

type ValidationResult int

const (
	NO_SIGNATURE ValidationResult = iota
	INVALID_FORMAT
	UNSUPPORTED_ALGORITHM
	MATCH
	MISMATCH
)

func (result ValidationResult) String() string {
	return strconv.Itoa(int(result))
}

func ValidateRequest(request *http.Request, key string) (
	result ValidationResult, headerSignature, computedSignature string) {
	headerSignature = request.Header.Get("Gap-Signature")
	if headerSignature == "" {
		result = NO_SIGNATURE
		return
	}

	components := strings.Split(headerSignature, " ")
	if len(components) != 2 {
		result = INVALID_FORMAT
		return
	}

	algorithm, err := HashAlgorithm(components[0])
	if err != nil {
		result = UNSUPPORTED_ALGORITHM
		return
	}

	computedSignature = RequestSignature(request, algorithm, key)
	if hmac.Equal([]byte(headerSignature), []byte(computedSignature)) {
		result = MATCH
	} else {
		result = MISMATCH
	}
	return
}
