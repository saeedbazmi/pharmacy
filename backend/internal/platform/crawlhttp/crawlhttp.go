// Package crawlhttp is the shared outbound HTTP layer for fetchers: timeout,
// transparent User-Agent, per-host rate limit, robots.txt and SSRF checks.
package crawlhttp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	userAgent = "PharmacyPriceCompare/0.1 (+https://github.com/saeedbazmi/pharmacy; price-comparison crawler)"
	maxBody   = 4 << 20
)

// Client is safe for concurrent use.
type Client struct {
	http     *http.Client
	minGap   time.Duration
	mu       sync.Mutex
	last     map[string]time.Time
	robots   map[string]robotsFile
	robotsAt map[string]time.Time
}

func New() *Client {
	return NewWithTimeout(15 * time.Second)
}

// NewWithTimeout builds a client whose every outbound call shares the given timeout.
func NewWithTimeout(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		http: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				if err := ValidateURL(req.URL.String()); err != nil {
					return err
				}
				return rejectPrivateHost(req.Context(), req.URL.Hostname())
			},
		},
		minGap:   1500 * time.Millisecond,
		last:     make(map[string]time.Time),
		robots:   make(map[string]robotsFile),
		robotsAt: make(map[string]time.Time),
	}
}

// Get fetches url after SSRF, robots and rate-limit checks.
func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, error) {
	if err := ValidateURL(rawURL); err != nil {
		return nil, err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if err := rejectPrivateHost(ctx, u.Hostname()); err != nil {
		return nil, err
	}
	if err := c.allowRobots(ctx, u); err != nil {
		return nil, err
	}
	if err := c.waitHost(ctx, u.Hostname()); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/json;q=0.9,*/*;q=0.8")

	const maxTries = 3
	var last error
	for attempt := 1; attempt <= maxTries; attempt++ {
		if attempt > 1 {
			if err := c.waitHost(ctx, u.Hostname()); err != nil {
				return nil, err
			}
			if err := sleepBackoff(ctx, attempt-1); err != nil {
				return nil, err
			}
		}
		body, err := c.doGet(req)
		if err == nil {
			return body, nil
		}
		last = err
		if !IsTransient(err) || attempt == maxTries {
			return nil, err
		}
	}
	return nil, last
}

func (c *Client) doGet(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, Transient(fmt.Errorf("get %s: %w", req.URL.Host, err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, Transient(fmt.Errorf("read body: %w", err))
	}
	if len(body) > maxBody {
		return nil, Structural(fmt.Errorf("response from %s exceeds %d bytes", req.URL.Host, maxBody))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}
	if isTransientStatus(resp.StatusCode) {
		return nil, Transient(fmt.Errorf("get %s: status %d", req.URL.Path, resp.StatusCode))
	}
	return nil, Structural(fmt.Errorf("get %s: status %d", req.URL.Path, resp.StatusCode))
}

func sleepBackoff(ctx context.Context, attempt int) error {
	d := 400 * time.Millisecond * time.Duration(1<<min(uint(attempt-1), 3))
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isTransientStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusRequestTimeout || code >= 500
}

func (c *Client) waitHost(ctx context.Context, host string) error {
	c.mu.Lock()
	last := c.last[host]
	wait := c.minGap - time.Since(last)
	c.mu.Unlock()
	if wait <= 0 {
		c.mu.Lock()
		c.last[host] = time.Now()
		c.mu.Unlock()
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		c.mu.Lock()
		c.last[host] = time.Now()
		c.mu.Unlock()
		return nil
	}
}

func (c *Client) allowRobots(ctx context.Context, u *url.URL) error {
	host := u.Hostname()
	c.mu.Lock()
	file, ok := c.robots[host]
	fetched := c.robotsAt[host]
	c.mu.Unlock()
	if !ok || time.Since(fetched) > time.Hour {
		var err error
		file, err = c.fetchRobots(ctx, u)
		if err != nil {
			// A missing robots.txt is treated as allow, per common crawler practice.
			file = robotsFile{allowAll: true}
		}
		c.mu.Lock()
		c.robots[host] = file
		c.robotsAt[host] = time.Now()
		c.mu.Unlock()
	}
	if !file.Allows(u.Path) {
		return Structural(fmt.Errorf("robots.txt disallows %s", u.Path))
	}
	return nil
}

func (c *Client) fetchRobots(ctx context.Context, page *url.URL) (robotsFile, error) {
	robotsURL := &url.URL{Scheme: page.Scheme, Host: page.Host, Path: "/robots.txt"}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL.String(), nil)
	if err != nil {
		return robotsFile{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return robotsFile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return robotsFile{allowAll: true}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return robotsFile{}, fmt.Errorf("robots.txt status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return robotsFile{}, err
	}
	return parseRobots(string(body)), nil
}

// ValidateURL rejects anything that is not public http(s). Used to prevent SSRF.
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme %q is not allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url is missing a host")
	}
	if ip := net.ParseIP(host); ip != nil && forbiddenIP(ip) {
		return fmt.Errorf("url host %s is not a public address", host)
	}
	return nil
}

func rejectPrivateHost(ctx context.Context, host string) error {
	if ip := net.ParseIP(host); ip != nil {
		if forbiddenIP(ip) {
			return fmt.Errorf("url host %s is not a public address", host)
		}
		return nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", host, err)
	}
	for _, rec := range ips {
		if forbiddenIP(rec.IP) {
			return fmt.Errorf("url host %s resolves to a private address", host)
		}
	}
	return nil
}

func forbiddenIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

type robotsFile struct {
	allowAll bool
	disallow []string
}

func parseRobots(body string) robotsFile {
	// Applies the * user-agent group only. Darukade currently Allows: /.
	inStar := false
	var disallow []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "user-agent:"):
			ua := strings.TrimSpace(line[len("user-agent:"):])
			inStar = ua == "*"
		case inStar && strings.HasPrefix(lower, "disallow:"):
			path := strings.TrimSpace(line[len("disallow:"):])
			if path != "" && path != "/" {
				disallow = append(disallow, path)
			}
		case inStar && strings.HasPrefix(lower, "allow:"):
			// Allow: / means nothing is disallowed by that rule.
		}
	}
	return robotsFile{disallow: disallow}
}

func (r robotsFile) Allows(path string) bool {
	if r.allowAll {
		return true
	}
	for _, prefix := range r.disallow {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}
	return true
}
