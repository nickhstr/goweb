package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rs/dnscache"
)

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
