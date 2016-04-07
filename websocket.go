package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ConnectionHeaderKey = http.CanonicalHeaderKey("connection")
	SetCookieHeaderKey  = http.CanonicalHeaderKey("set-cookie")
	UpgradeHeaderKey    = http.CanonicalHeaderKey("upgrade")
	WSKeyHeaderKey      = http.CanonicalHeaderKey("sec-websocket-key")
	WSProtocolHeaderKey = http.CanonicalHeaderKey("sec-websocket-protocol")
	WSVersionHeaderKey  = http.CanonicalHeaderKey("sec-websocket-version")

	ConnectionHeaderValue = "Upgrade"
	UpgradeHeaderValue    = "websocket"

	HandshakeHeaders = []string{ConnectionHeaderKey, UpgradeHeaderKey, WSVersionHeaderKey, WSKeyHeaderKey}
	UpgradeHeaders   = []string{SetCookieHeaderKey, WSProtocolHeaderKey}
)

func (u *UpstreamProxy) handleWebsocket(w http.ResponseWriter, r *http.Request) {

	// Copy request headers and remove websocket handshaking headers
	// before submitting to the upstream server
	upstreamHeader := http.Header{}
	for key, _ := range r.Header {
		copyHeader(&upstreamHeader, r.Header, key)
	}
	for _, header := range HandshakeHeaders {
		delete(upstreamHeader, header)
	}
	upstreamHeader.Set("Host", r.Host)

	// Connect upstream
	upstreamAddr := u.upstreamWSURL(*r.URL).String()
	upstream, upstreamResp, err := websocket.DefaultDialer.Dial(upstreamAddr, upstreamHeader)
	if err != nil {
		if upstreamResp != nil {
			log.Printf("dialing upstream websocket failed with code %d: %v", upstreamResp.StatusCode, err)
		} else {
			log.Printf("dialing upstream websocket failed: %v", err)
		}
		http.Error(w, "websocket unavailable", http.StatusServiceUnavailable)
		return
	}
	defer upstream.Close()

	// Pass websocket handshake response headers to the upgrader
	upgradeHeader := http.Header{}
	copyHeaders(&upgradeHeader, upstreamResp.Header, UpgradeHeaders)

	// Upgrade the client connection without validating the origin
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	client, err := upgrader.Upgrade(w, r, upgradeHeader)
	if err != nil {
		log.Printf("couldn't upgrade websocket request: %v", err)
		http.Error(w, "websocket upgrade failed", http.StatusServiceUnavailable)
		return
	}

	// Wire both sides together and close when finished
	var wg sync.WaitGroup
	cp := func(dst, src *websocket.Conn) {
		defer wg.Done()
		_, err := io.Copy(dst.UnderlyingConn(), src.UnderlyingConn())

		var closeMessage []byte
		if err != nil {
			closeMessage = websocket.FormatCloseMessage(websocket.CloseProtocolError, err.Error())
		} else {
			closeMessage = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye")
		}
		// Attempt to close the connection properly
		dst.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(2*time.Second))
		src.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(2*time.Second))
	}
	wg.Add(2)
	go cp(upstream, client)
	go cp(client, upstream)
	wg.Wait()
}

// Create a websocket URL from the request URL
func (u *UpstreamProxy) upstreamWSURL(r url.URL) *url.URL {
	ws := r
	ws.User = r.User
	ws.Host = u.upstream.Host
	ws.Fragment = ""
	switch u.upstream.Scheme {
	case "http":
		ws.Scheme = "ws"
	case "https":
		ws.Scheme = "wss"
	}
	return &ws
}

func isWebsocketRequest(req *http.Request) bool {
	return isHeaderValuePresent(req.Header, UpgradeHeaderKey, UpgradeHeaderValue) &&
		isHeaderValuePresent(req.Header, ConnectionHeaderKey, ConnectionHeaderValue)
}

func isHeaderValuePresent(headers http.Header, key string, value string) bool {
	for _, header := range headers[key] {
		for _, v := range strings.Split(header, ",") {
			if strings.EqualFold(value, strings.TrimSpace(v)) {
				return true
			}
		}
	}
	return false
}

func copyHeaders(dst *http.Header, src http.Header, headers []string) {
	for _, header := range headers {
		copyHeader(dst, src, header)
	}
}

// Copy any non-empty and non-blank header values
func copyHeader(dst *http.Header, src http.Header, header string) {
	for _, value := range src[header] {
		if value != "" {
			dst.Add(header, value)
		}
	}
}
