package client

import (
	"net"
	"net/http"
	"time"

	"github.com/cockroachdb/cmux"
)

type Server struct {
	Addr       string
	Handler    http.Handler
	ServerAddr string
}

func (s *Server) Serve(l net.Listener) error {
	m := cmux.New(l)
	websocketL := m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))
	httpL := m.Match(cmux.Any())

	go func() {
		s := &http.Server{
			Addr:           s.Addr,
			Handler:        s.Handler,
			ReadTimeout:    5 * time.Minute,
			WriteTimeout:   5 * time.Minute,
			MaxHeaderBytes: 1 << 20,
		}
		if err := s.Serve(httpL); err != cmux.ErrListenerClosed {
			panic(err)
		}
	}()

	host, port, err := net.SplitHostPort(s.ServerAddr)
	if err != nil {
		return err
	}
	proxyTarget := s.ServerAddr
	if net.ParseIP(host).IsUnspecified() {
		// handle case where host is 0.0.0.0
		proxyTarget = net.JoinHostPort("127.0.0.1", port)
	}
	wsproxy, err := NewProxy(proxyTarget)
	if err != nil {
		return err
	}

	go func() {
		s := &http.Server{
			Handler: wsproxy,
		}
		if err := s.Serve(websocketL); err != cmux.ErrListenerClosed {
			panic(err)
		}
	}()

	return m.Serve()
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}
