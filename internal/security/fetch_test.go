package security

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"testing"
)

type rt func(*http.Request) (*http.Response, error)

func (f rt) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestSSRFAddresses(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "100.100.100.200", "169.254.169.254", "192.168.1.1", "::1", "::ffff:127.0.0.1", "fc00::1", "fe80::1", "2002:7f00:1::", "64:ff9b::7f00:1"} {
		if PublicIP(netip.MustParseAddr(ip)) {
			t.Fatal(ip)
		}
	}
	for _, u := range []string{"file:///etc/passwd", "http://localhost:9000/", "https://user:pass@example.org/", "http://127.0.0.1/", "https://[::1]/"} {
		if _, e := ValidateURL(u); e == nil {
			t.Fatal(u)
		}
	}
	if !PublicIP(netip.MustParseAddr("1.1.1.1")) {
		t.Fatal("public blocked")
	}
}
func response(status int, body string, headers http.Header) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: headers}
}
func TestRedirectRobotsAndLimits(t *testing.T) {
	f := NewFetcher()
	f.client.Transport = rt(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/robots.txt" {
			return response(200, "User-agent: *\nDisallow: /private", http.Header{"Content-Type": {"text/plain"}}), nil
		}
		if r.URL.Path == "/redirect" {
			return response(302, "", http.Header{"Location": {"http://169.254.169.254/latest"}}), nil
		}
		return response(200, strings.Repeat("x", 100), http.Header{"Content-Type": {"text/html"}}), nil
	})
	if _, _, e := f.Fetch(context.Background(), "https://example.org/private", 1000); e == nil {
		t.Fatal("robots ignored")
	}
	if _, _, e := f.Fetch(context.Background(), "https://example.org/redirect", 1000); e == nil {
		t.Fatal("unsafe redirect")
	}
	if _, _, e := f.Fetch(context.Background(), "https://example.org/ok", 10); e == nil {
		t.Fatal("limit ignored")
	}
	if _, _, e := f.Fetch(context.Background(), "https://example.org/ok", 1000); e != nil {
		t.Fatal(e)
	}
}

func TestDNSValidationAndPinnedDial(t *testing.T) {
	calls := 0
	lookup := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("127.0.0.1")}, nil
	}
	dial := func(context.Context, string, string) (net.Conn, error) {
		calls++
		return nil, errors.New("unexpected dial")
	}
	if _, err := dialPublic(context.Background(), "tcp", "example.org:443", lookup, dial); err == nil || calls != 0 {
		t.Fatal("mixed public/private DNS accepted")
	}
	resolved := 0
	lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		resolved++
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	}
	dial = func(_ context.Context, network, address string) (net.Conn, error) {
		if address != "1.1.1.1:443" {
			t.Fatal("hostname re-resolved instead of pinned", address)
		}
		a, b := net.Pipe()
		b.Close()
		return a, nil
	}
	conn, err := dialPublic(context.Background(), "tcp", "example.org:443", lookup, dial)
	if err != nil || resolved != 1 {
		t.Fatal(err, resolved)
	}
	conn.Close()
}
