package httpserver

import (
	"fmt"
	"net"
	"net/http"
)

func ListenAndServe(addr string, handler http.Handler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("HTTP server listened on %s\n", ln.Addr())

	server := &http.Server{Addr: addr, Handler: handler}
	return server.Serve(ln)
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("HTTPs server listened on %s\n", ln.Addr())

	server := &http.Server{Addr: addr, Handler: handler}
	return server.ServeTLS(ln, certFile, keyFile)
}
