package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nickhstr/goweb/env"
	"github.com/rs/dnscache"
	"github.com/rs/zerolog/log"
)

// Start creates and starts a server, listening on "address"
func Start(mux http.Handler) {
	var (
		address  string
		listener net.Listener
		mode     = env.Get("GO_ENV", "development")
		err      error
	)

	address = net.JoinHostPort(Host(), Port())
	listener, err = PreferredListener(address)
	if err != nil {
		// Non-nil error means the address wanted is taken. Time to find a free one.
		listener = FreePortListener()
		if listener != nil {
			address = listener.Addr().String()
		}
	}

	dnsCacheEnabled, err := strconv.ParseBool(env.Get("DNS_CACHE_ENABLED", "true"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to convert 'DNS_CACHE_ENABLED' to bool")
	}
	// TTL measured in seconds
	dnsCacheTTTL, err := strconv.Atoi(env.Get("DNS_CACHE_TTL", "300"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to convert 'DNS_CACHE_TTL' to int")
	}
	DNSCache(
		dnsCacheEnabled,
		dnsCacheTTTL,
	)

	srv := &http.Server{Handler: mux}

	idlConnsClosed := make(chan struct{})
	go shutdown(srv, idlConnsClosed)

	log.Log().
		Str("address", address).
		Str("mode", mode).
		Msg("Server listening")

	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Server failed to start")
	}

	<-idlConnsClosed
}

// Shutdown server gracefully on SIGINT or SIGTERM
func shutdown(server *http.Server, idleConnectionsClosed chan struct{}) {
	// Block until signal is received
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("Server shutdown error")
	}

	log.Log().Msg("Server shutdown")

	// Close channel to signal shutdown is complete
	close(idleConnectionsClosed)
}

// DNSCache adds caching to dns lookups, performed by the standard library's http.DefaultClient.
// TTL in this case doesn't truly mean TTL for an address; rather, it determines the number of
// seconds to use for the cache refresh interval.
func DNSCache(enable bool, ttl int) {
	if !enable {
		return
	}

	r := &dnscache.Resolver{}
	http.DefaultTransport.(*http.Transport).
		DialContext = func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
		separator := strings.LastIndex(addr, ":")
		ips, err := r.LookupHost(ctx, addr[:separator])
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			conn, err = net.Dial(network, ip+addr[separator:])
			if err == nil {
				break
			}
		}
		return
	}

	// Run refresh job in background
	go func() {
		t := time.NewTicker(time.Duration(ttl) * time.Second)
		defer t.Stop()
		for range t.C {
			// Use true to refresh addresses not used since the last refresh
			r.Refresh(true)
		}
	}()
}

// PreferredListener will attempt to create a listener for the given address.
func PreferredListener(addr string) (net.Listener, error) {
	var (
		listener net.Listener
		err      error
	)

	if addr == "" {
		return nil, errors.New("Address must be a non-empty string")
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
func FreePortListener() net.Listener {
	listener, err := net.Listen("tcp", net.JoinHostPort(Host(), "0"))
	if err != nil {
		// If this can't find a free port, we need to start panicking!
		panic(err)
	}

	return listener
}

// Host gets the host for a listener's address.
func Host() string {
	var defaultHost string

	if env.Dev() {
		defaultHost = "localhost"
	} else {
		defaultHost = "0.0.0.0"
	}

	return env.Get("HOST", defaultHost)
}

// Port gets the port for a listener's address.
func Port() string {
	return env.Get("PORT", "3000")
}
