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
	assert.Equal(t, expected, dst)
}

func TestUpgrade(t *testing.T) {
	tests := []struct {
		upgrade         bool
		connectionValue string
		upgradeValue    string
	}{
		{true, "Upgrade", "Websocket"},
		{true, "keepalive, Upgrade", "websocket"},
		{false, "", "websocket"},
		{false, "keepalive, Upgrade", ""},
	}

	for _, tt := range tests {
		req := new(http.Request)
		req.Header = http.Header{}
		req.Header.Set(ConnectionHeaderKey, tt.connectionValue)
		req.Header.Set(UpgradeHeaderKey, tt.upgradeValue)
		assert.Equal(t, tt.upgrade, isWebsocketRequest(req))
	}
}
