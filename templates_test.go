package main

import (
	"github.com/bitly/oauth2_proxy/Godeps/_workspace/src/github.com/bmizerany/assert"
	"testing"
)

func TestTemplatesCompile(t *testing.T) {
	templates := getTemplates()
	assert.NotEqual(t, templates, nil)
}
