package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type WebsocketReverseProxy struct {
	upstream string
	handler  http.Handler
}

func (p *WebsocketReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if websocketUpgradeRequest(req) {
		p.hijackWebsocket(rw, req)
	} else {
		p.handler.ServeHTTP(rw, req)
	}
}

func (p *WebsocketReverseProxy) hijackWebsocket(rw http.ResponseWriter, req *http.Request) {
	highjacker, ok := rw.(http.Hijacker)

	if !ok {
		log.Printf("webserver doesn't support hijacking")
		http.Error(rw, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	conn, _, err := highjacker.Hijack()
	if err != nil {
		log.Printf("failed to hijack connection: %v", err)
		http.Error(rw, "failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	conn2, err := net.Dial("tcp", p.upstream)
	if err != nil {
		log.Printf("couldn't connect to backend websocket server: %v", err)
		http.Error(rw, "couldn't connect to backend server", http.StatusServiceUnavailable)
		return
	}
	defer conn2.Close()

	err = req.Write(conn2)
	if err != nil {
		log.Printf("writing WebSocket request to backend server failed: %v", err)
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(conn2, conn)
	go cp(conn, conn2)

	err = <-errc
	if err != nil {
		log.Printf("error copying data: %v", err)
		return
	}
}

func websocketUpgradeRequest(req *http.Request) bool {
	connectionHeaders, ok := req.Header["Connection"]
	if !ok || len(connectionHeaders) <= 0 {
		return false
	}

	connectionHeader := connectionHeaders[0]
	if strings.ToLower(connectionHeader) != "upgrade" {
		return false
	}

	upgradeHeaders, ok := req.Header["Upgrade"]
	if !ok || len(upgradeHeaders) <= 0 {
		return false
	}

	return strings.ToLower(upgradeHeaders[0]) == "websocket"
}
