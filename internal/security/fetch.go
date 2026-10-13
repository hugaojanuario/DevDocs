package security

import (
	"context"
	"errors"
	"fmt"
	"github.com/temoto/robotstxt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"
)

const UserAgent = "DevDocs/1.0 (+documentation research; respects robots.txt)"

var ErrBlocked = errors.New("destination blocked")
var forbidden = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"), netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("::/96"), netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("64:ff9b:1::/48"), netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("3fff::/20"), netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
}

func PublicIP(a netip.Addr) bool {
	a = a.Unmap()
	if !a.IsGlobalUnicast() || (a.Is6() && !netip.MustParsePrefix("2000::/3").Contains(a)) {
		return false
	}
	for _, p := range forbidden {
		if p.Contains(a) {
			return false
		}
	}
	return true
}
func ValidateURL(raw string) (*url.URL, error) {
	u, e := url.Parse(raw)
	if e != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrBlocked
	}
	p := u.Port()
	if p != "" && !(u.Scheme == "http" && p == "80") && !(u.Scheme == "https" && p == "443") {
		return nil, ErrBlocked
	}
	if a, e := netip.ParseAddr(u.Hostname()); e == nil && !PublicIP(a) {
		return nil, ErrBlocked
	}
	return u, nil
}

type Fetcher struct {
	client *http.Client
	slots  chan struct{}
	mu     sync.Mutex
	robots map[string]robotEntry
}
type robotEntry struct {
	data    *robotstxt.RobotsData
	expires time.Time
	next    time.Time
}

func NewFetcher() *Fetcher {
	t := &http.Transport{Proxy: nil, MaxIdleConns: 16, MaxIdleConnsPerHost: 2, IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 8 * time.Second}
	t.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		d := &net.Dialer{Timeout: 5 * time.Second}
		return dialPublic(ctx, network, address, net.DefaultResolver.LookupNetIP, d.DialContext)
	}

	return &Fetcher{client: &http.Client{Transport: t, Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, slots: make(chan struct{}, 4), robots: map[string]robotEntry{}}
}
func (f *Fetcher) request(ctx context.Context, raw string, max int64, headers http.Header) ([]byte, int, string, string, error) {
	u, e := ValidateURL(raw)
	if e != nil {
		return nil, 0, "", "", e
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return nil, 0, "", "", e
	}
	req.Header.Set("User-Agent", UserAgent)
	for k, v := range headers {
		req.Header[k] = v
	}
	res, e := f.client.Do(req)
	if e != nil {
		return nil, 0, "", "", fmt.Errorf("source request unavailable")
	}
	defer res.Body.Close()
	b, e := io.ReadAll(io.LimitReader(res.Body, max+1))
	if e != nil || int64(len(b)) > max {
		return nil, res.StatusCode, "", "", errors.New("content limit exceeded")
	}
	return b, res.StatusCode, res.Header.Get("Content-Type"), res.Header.Get("Location"), nil
}
func (f *Fetcher) robotsAllowed(ctx context.Context, u *url.URL) error {
	origin := u.Scheme + "://" + u.Host
	f.mu.Lock()
	entry, ok := f.robots[origin]
	f.mu.Unlock()
	if !ok || time.Now().After(entry.expires) {
		raw := origin + "/robots.txt"
		var b []byte
		var status int
		var e error
		for i := 0; i < 4; i++ {
			var loc string
			b, status, _, loc, e = f.request(ctx, raw, 512<<10, nil)
			if e != nil {
				return errors.New("robots unavailable")
			}
			if status >= 300 && status < 400 {
				base, _ := url.Parse(raw)
				next, err := base.Parse(loc)
				if err != nil || loc == "" {
					return ErrBlocked
				}
				raw = next.String()
				continue
			}
			break
		}
		var data *robotstxt.RobotsData
		if status == 404 || status == 410 {
			data, e = robotstxt.FromBytes(nil)
		} else if status >= 200 && status < 300 {
			data, e = robotstxt.FromBytes(b)
		} else {
			return errors.New("robots unavailable or access denied")
		}
		if e != nil {
			return errors.New("invalid robots")
		}
		entry = robotEntry{data: data, expires: time.Now().Add(5 * time.Minute)}
	}
	group := entry.data.FindGroup("DevDocs")
	if !group.Test(u.RequestURI()) {
		return errors.New("robots disallows source")
	}
	delay := group.CrawlDelay
	if delay < 250*time.Millisecond {
		delay = 250 * time.Millisecond
	}
	if delay > 5*time.Second {
		return errors.New("source crawl delay exceeds budget")
	}
	f.mu.Lock()
	if existing, ok := f.robots[origin]; ok && existing.next.After(entry.next) {
		entry.next = existing.next
	}
	start := entry.next
	if start.Before(time.Now()) {
		start = time.Now()
	}
	entry.next = start.Add(delay)
	if len(f.robots) >= 64 {
		for k, v := range f.robots {
			if time.Now().After(v.expires) {
				delete(f.robots, k)
			}
		}
		if len(f.robots) >= 64 {
			for k := range f.robots {
				delete(f.robots, k)
				break
			}
		}
	}
	f.robots[origin] = entry
	f.mu.Unlock()
	timer := time.NewTimer(time.Until(start))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (f *Fetcher) Fetch(ctx context.Context, raw string, max int64) ([]byte, string, error) {
	return f.get(ctx, raw, max, true, nil)
}

// Provider APIs are authenticated contractual endpoints; never leak credentials on redirects.
func (f *Fetcher) Provider(ctx context.Context, raw string, headers http.Header) ([]byte, error) {
	b, _, e := f.get(ctx, raw, 2<<20, false, headers)
	return b, e
}
func (f *Fetcher) get(ctx context.Context, raw string, max int64, robots bool, headers http.Header) ([]byte, string, error) {
	select {
	case f.slots <- struct{}{}:
		defer func() { <-f.slots }()
	case <-ctx.Done():
		return nil, "", ctx.Err()
	}
	for i := 0; i < 5; i++ {
		u, e := ValidateURL(raw)
		if e != nil {
			return nil, "", e
		}
		if robots {
			if e = f.robotsAllowed(ctx, u); e != nil {
				return nil, "", e
			}
		}
		b, status, typ, loc, e := f.request(ctx, raw, max, headers)
		if e != nil {
			return nil, "", e
		}
		if status >= 300 && status < 400 {
			if !robots {
				return nil, "", errors.New("provider redirect rejected")
			}
			next, e := u.Parse(loc)
			if e != nil || loc == "" {
				return nil, "", ErrBlocked
			}
			next.Fragment = ""
			if u.Scheme == "https" && next.Scheme != "https" {
				return nil, "", ErrBlocked
			}
			raw = next.String()
			continue
		}
		if status != 200 {
			return nil, "", fmt.Errorf("source returned HTTP %d", status)
		}
		media, _, _ := mime.ParseMediaType(typ)
		if robots && media != "text/html" && media != "application/xhtml+xml" && !strings.Contains(media, "javascript") && media != "application/json" && media != "text/plain" && media != "application/xml" && media != "text/xml" {
			return nil, "", errors.New("unsupported source content type")
		}
		return b, raw, nil
	}
	return nil, "", errors.New("redirect limit exceeded")
}

func dialPublic(ctx context.Context, network, address string, lookup func(context.Context, string, string) ([]netip.Addr, error), dial func(context.Context, string, string) (net.Conn, error)) (net.Conn, error) {
	h, p, e := net.SplitHostPort(address)
	if e != nil {
		return nil, ErrBlocked
	}
	ips, e := lookup(ctx, "ip", h)
	if e != nil || len(ips) == 0 {
		return nil, ErrBlocked
	}
	for _, ip := range ips {
		if !PublicIP(ip) {
			return nil, ErrBlocked
		}
	}
	for _, ip := range ips {
		c, e := dial(ctx, network, net.JoinHostPort(ip.String(), p))
		if e == nil {
			return c, nil
		}
	}
	return nil, ErrBlocked
}
