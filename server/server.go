package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("server")

// Start creates and starts a server, listening on "address"
func Start(mux http.Handler) error {
	// Default server timout, in seconds
	const defaultTimeout = 15
	var (
		address  string
		listener net.Listener
		err      error
	)

	address = net.JoinHostPort(Host(), env.Get("PORT", "3000"))
	listener, err = PreferredListener(address)
	if err != nil {
		// Non-nil error means the address wanted is taken. Time to find a free one.
		listener, err = FreePortListener()
		if err != nil {
			return err
		}
		if listener != nil {
			address = listener.Addr().String()
		}
	}

	dnsCacheEnabled, err := strconv.ParseBool(env.Get("DNS_CACHE_ENABLED", "true"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to convert 'DNS_CACHE_ENABLED' to bool")
		return err
	}
	// TTL measured in seconds
	dnsCacheTTL, err := strconv.Atoi(env.Get("DNS_CACHE_TTL", "300"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to convert 'DNS_CACHE_TTL' to int")
		return err
	}
	DNSCache(
		dnsCacheEnabled,
		dnsCacheTTL,
	)

	srvTimeout, err := strconv.Atoi(env.Get("SERVER_TIMEOUT", strconv.Itoa(defaultTimeout)))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Invalid SERVER_TIMEOUT set")
		return err
	}

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  time.Duration(srvTimeout) * time.Second,
		WriteTimeout: time.Duration(srvTimeout) * time.Second,
	}

	idlConnsClosed := make(chan struct{})
	go shutdown(srv, idlConnsClosed)

	log.Log().
		Str("address", address).
		Str("mode", env.Get("GO_ENV", "development")).
		Msg("Server listening")

	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Server failed to start")
		return err
	}

	<-idlConnsClosed

	return nil
}

// Shutdown server gracefully on SIGINT or SIGTERM
func shutdown(srv *http.Server, idleConnectionsClosed chan struct{}) {
	// Block until signal is received
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("Server shutdown error")
	}

	log.Log().Msg("Server shutdown")

	// Close channel to signal shutdown is complete
	close(idleConnectionsClosed)
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
