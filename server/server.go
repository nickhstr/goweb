package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("server")

// Server is a light wrapper around http.Server
type Server struct {
	*http.Server
	address  string
	listener net.Listener
	// Channel used to signal server has shutdown
	serverShutdown chan struct{}
}

// StartNew creates a Server and starts it.
func StartNew(h http.Handler) error {
	srv, err := New(&http.Server{
		Handler: h,
	})
	if err != nil {
		return err
	}

	return srv.Start()
}

// New creates a new Server.
func New(s *http.Server) (*Server, error) {
	// Default server timout, in seconds
	const defaultSrvTimeout = 15 * time.Second
	var (
		srv *Server
		err error
	)

	address := net.JoinHostPort(Host(), env.Get("PORT", "3000"))
	listener, err := PreferredListener(address)
	if err != nil {
		// Non-nil error means the address wanted is taken. Time to find a free one.
		listener, err = FreePortListener()
		if err != nil {
			return nil, err
		}
		if listener != nil {
			address = listener.Addr().String()
		}
	}

	// ensure timeouts are set
	if s.ReadTimeout == 0 {
		s.ReadTimeout = defaultSrvTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = defaultSrvTimeout
	}

	srv = &Server{
		s,
		address,
		listener,
		make(chan struct{}),
	}

	return srv, nil
}

// Start begins serving, and listens for termination signals to shutdown gracefully.
func (srv *Server) Start() error {
	var err error

	go srv.shutdown()

	log.Log().
		Str("address", srv.address).
		Str("mode", env.Get("GO_ENV", "development")).
		Msg("Server listening")

	err = srv.Serve(srv.listener)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Server failed to start")
		return err
	}

	<-srv.serverShutdown

	return nil
}

// Shutdown server gracefully on SIGINT or SIGTERM.
func (srv *Server) shutdown() {
	// Block until signal is received
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("Server shutdown error")
	}

	log.Log().Msg("Server shutdown")

	// Close channel to signal shutdown is complete
	close(srv.serverShutdown)
}

// PreferredListener will attempt to create a listener for the given address.
func PreferredListener(addr string) (net.Listener, error) {
	var (
		listener net.Listener
		err      error
	)

	if addr == "" {
		return nil, errors.New("address must be a non-empty string")
	}

	listener, err = net.Listen("tcp", addr)
	if listener != nil {
		return listener, nil
	} else if err != nil {
		log.Error().Err(err).Msgf("%s unavailable", addr)
	}

	return listener, err
}

// FreePortListener will return a listener for any available port on the Host.
func FreePortListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(Host(), "0"))
	if err != nil {
		return nil, fmt.Errorf("FreePortListener could not find free port: %w", err)
	}

	return listener, nil
}

// Host gets the host for a listener's address.
func Host() string {
	var defaultHost string

	if env.IsDev() {
		defaultHost = "localhost"
	} else {
		defaultHost = "0.0.0.0"
	}

	return env.Get("HOST", defaultHost)
}
