package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

type WebsocketReverseProxy struct {
	*httputil.ReverseProxy
	Upstream string
}

func NewWebsocketReverseProxy(target *url.URL) *WebsocketReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return &WebsocketReverseProxy{ReverseProxy: proxy, Upstream: target.Host}
}

func (p *WebsocketReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if websocketUpgradeRequest(req) {
		p.hijackWebsocket(rw, req)
	} else {
		p.ReverseProxy.ServeHTTP(rw, req)
	}
}

func (p *WebsocketReverseProxy) hijackWebsocket(rw http.ResponseWriter, req *http.Request) {
	highjacker, ok := rw.(http.Hijacker)

	if !ok {
		http.Error(rw, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := highjacker.Hijack()
	defer conn.Close()

	conn2, err := net.Dial("tcp", p.Upstream)
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

	bufferedBidirCopy(conn, bufrw, conn2, bufio.NewReadWriter(bufio.NewReader(conn2), bufio.NewWriter(conn2)))
}

func websocketUpgradeRequest(req *http.Request) bool {
	connection_headers, ok := req.Header["Connection"]
	if !ok || len(connection_headers) <= 0 {
		return false
	}

	connection_header := connection_headers[0]
	if strings.ToLower(connection_header) != "upgrade" {
		return false
	}

	upgrade_headers, ok := req.Header["Upgrade"]
	if !ok || len(upgrade_headers) <= 0 {
		return false
	}

	return strings.ToLower(upgrade_headers[0]) == "websocket"
}

func bufferedCopy(dest *bufio.ReadWriter, src *bufio.ReadWriter) {
	buf := make([]byte, 40*1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("Upstream read failed: %v", err)
			return
		}
		if n == 0 {
			return
		}
		n, err = dest.Write(buf[0:n])
		if err != nil && err != io.EOF {
			log.Printf("Downstream write failed: %v", err)
			return
		}

		err = dest.Flush()
		if err != nil {
			log.Printf("Downstream write flush failed: %v", err)
			return
		}
	}
}

func bufferedBidirCopy(conn1 io.ReadWriteCloser, rw1 *bufio.ReadWriter, conn2 io.ReadWriteCloser, rw2 *bufio.ReadWriter) {
	wg := sync.WaitGroup{}

	copier := func(wg *sync.WaitGroup, rw1 *bufio.ReadWriter, rw2 *bufio.ReadWriter) {
		defer wg.Done()
		bufferedCopy(rw2, rw1)
	}

	wg.Add(2)
	go copier(&wg, rw1, rw2)
	go copier(&wg, rw2, rw1)
	wg.Wait()
}
