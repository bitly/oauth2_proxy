package main

import (
	"bytes"
	"github.com/bmizerany/assert"
	"testing"
)

func TestHtpasswdSha(t *testing.T) {
	file := bytes.NewBuffer([]byte("testuser:{SHA}PaVBVZkYqAjCQCu6UBL2xgsnZhw=\n"))
	h, err := NewHtpasswd(file)
	assert.Equal(t, err, nil)

	valid := h.Validate("testuser", "asdf")
	assert.Equal(t, valid, true)
}

func TestHtpasswdBcrypt(t *testing.T) {
	file := bytes.NewBuffer([]byte("testuser:$2y$05$38ykzs7.Z.C.9yaJSknchOw7hkMJmTVGZAIzV5ThjIWLXgOlCLWeC\n"))
	h, err := NewHtpasswd(file)
	assert.Equal(t, err, nil)

	valid := h.Validate("testuser", "asdf")
	assert.Equal(t, valid, true)
}
