// Package dnscache exports DNS caching utilities for http.DialContexts.
package dnscache

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rs/dnscache"
)

// Disable resets the default dial context back to http.Transport's DialContext.
func Disable() {
	// The default DialContext used by http.DefaultTransport
	defaultDialContext := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext
	http.DefaultTransport.(*http.Transport).DialContext = defaultDialContext
}

// Enable adds caching to DNS lookups, performed by the standard library's http.DefaultTransport.
// TTL in this case doesn't truly mean TTL for an address; rather, it determines the number of
// seconds to use for the cache refresh interval.
func Enable(ttl int) {
	dc := DialContext(ttl)
	http.DefaultTransport.(*http.Transport).DialContext = dc
}

// DialContext returns an http.DialContext with DNS caching.
// Cached DNS entries are refreshed by a ticker, with the given TTL as the
// number of seconds to wait before refreshing the cache.
func DialContext(ttl int) func(context.Context, string, string) (net.Conn, error) {
	r := new(dnscache.Resolver)

	// Run refresh job in background
	go func() {
		t := time.NewTicker(time.Duration(ttl) * time.Second)
		defer t.Stop()

		for range t.C {
			// Use true to refresh addresses not used since the last refresh
			r.Refresh(true)
		}
	}()

	return cachedDialContext(r)
}

// cachedDialContext returns a DialContext which caches DNS lookups of HTTP connections.
func cachedDialContext(r *dnscache.Resolver) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		var (
			conn net.Conn
			err  error
		)

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		ips, err := r.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			var dialer net.Dialer

			conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				break
			}
		}

		return conn, err
	}
}
