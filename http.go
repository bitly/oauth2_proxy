package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	Handler http.Handler
	Opts    *Options
}

func (s *Server) ListenAndServe() {
	if s.Opts.TLSKeyFile != "" || s.Opts.TLSCertFile != "" {
		s.ServeHTTPS()
	} else {
		s.ServeHTTP()
	}
}

func (s *Server) ServeHTTP() {
	httpAddress := s.Opts.HttpAddress
	scheme := ""

	i := strings.Index(httpAddress, "://")
	if i > -1 {
		scheme = httpAddress[0:i]
	}

	var networkType string
	switch scheme {
	case "", "http":
		networkType = "tcp"
	default:
		networkType = scheme
	}

	slice := strings.SplitN(httpAddress, "//", 2)
	listenAddr := slice[len(slice)-1]

	listener, err := net.Listen(networkType, listenAddr)
	if err != nil {
		log.Fatalf("FATAL: listen (%s, %s) failed - %s", networkType, listenAddr, err)
	}
	log.Printf("HTTP: listening on %s", listenAddr)

	server := &http.Server{Handler: s.Handler}
	err = server.Serve(listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Printf("ERROR: http.Serve() - %s", err)
	}

	log.Printf("HTTP: closing %s", listener.Addr())
}

func (s *Server) ServeHTTPS() {
	addr := s.Opts.HttpsAddress
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(s.Opts.TLSCertFile, s.Opts.TLSKeyFile)
	if err != nil {
		log.Fatalf("FATAL: loading tls config (%s, %s) failed - %s", s.Opts.TLSCertFile, s.Opts.TLSKeyFile, err)
	}

	srv := &http.Server{
		Addr:      addr,
		Handler:   s.Handler,
		TLSConfig: config,
	}
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
