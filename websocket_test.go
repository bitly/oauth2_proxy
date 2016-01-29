package main

import (
	"net/http"
	"testing"

	"github.com/bmizerany/assert"
)

func TestCopyHeader(t *testing.T) {
	src := http.Header{
		"EmptyValue": []string{""},
		"Nil":        []string{},
		"Single":     []string{"one"},
		"Multi":      []string{"one", "two"},
	}
	expected := http.Header{
		"Single": []string{"one"},
		"Multi":  []string{"one", "two"},
	}
	dst := http.Header{}
	for key, _ := range src {
		copyHeader(&dst, src, key)
	}
	assert.Equal(t, dst, expected)
}
