// Package server provides an enhanced http.Server and convenience functions.
// Servers are designed to be robust, flexible, and graceful in their shutdown
// process.
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

	"github.com/nickhstr/goweb/config"
	"github.com/nickhstr/goweb/logger"
	"github.com/spf13/viper"
)

var log = logger.New("server")

// Server is a light wrapper around http.Server.
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

// New creates a new Server, built off of a base http.Server.
func New(s *http.Server) (*Server, error) {
	// Default server timout, in seconds
	const defaultSrvTimeout = 15 * time.Second

	var (
		srv *Server
		err error
	)

	port := viper.GetString("PORT")
	if port == "" {
		port = "3000"
	}

	address := net.JoinHostPort(Host(), port)
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
		Str("mode", viper.GetString("GO_ENV")).
		Int("pid", os.Getpid()).
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

	// Allow up to thirty seconds for server operations to finish before
	// canceling them.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
	host := viper.GetString("HOST")
	if host != "" {
		return host
	}

	if config.IsDev() {
		host = "localhost"
	} else {
		host = "0.0.0.0"
	}

	return host
}
