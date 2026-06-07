package urlguard

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPURLOptions controls how an HTTP(S) URL is normalized before storage or use.
type HTTPURLOptions struct {
	AllowEmpty         bool
	AllowPrivateHosts  bool
	StripTrailingSlash bool
	RejectQuery        bool
	RejectFragment     bool
}

// NormalizeHTTPURL trims and normalizes an absolute http/https URL.
func NormalizeHTTPURL(rawURL string, opts HTTPURLOptions) (string, error) {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		if opts.AllowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid url format: %w", err)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("url must use http or https")
	}
	if parsed.Host == "" || parsed.Hostname() == "" {
		return "", fmt.Errorf("url must have a host")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("url must not contain user info")
	}
	if !opts.AllowPrivateHosts && isPrivateOrLocalHost(parsed.Hostname()) {
		return "", fmt.Errorf("url must not point to localhost or private network")
	}
	if opts.RejectQuery && parsed.RawQuery != "" {
		return "", fmt.Errorf("url must not contain query parameters")
	}
	if opts.RejectFragment && parsed.Fragment != "" {
		return "", fmt.Errorf("url must not contain a fragment")
	}

	parsed.Scheme = scheme
	if opts.StripTrailingSlash {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
		parsed.RawPath = ""
	}
	parsed.Fragment = ""

	return parsed.String(), nil
}

// NormalizeServiceBaseURL normalizes an admin-trusted upstream service root.
func NormalizeServiceBaseURL(rawURL string) (string, error) {
	return NormalizeHTTPURL(rawURL, HTTPURLOptions{
		AllowPrivateHosts:  true,
		StripTrailingSlash: true,
		RejectQuery:        true,
		RejectFragment:     true,
	})
}

// NormalizeTrustedCallbackURL normalizes an admin-trusted callback URL.
func NormalizeTrustedCallbackURL(rawURL string, allowEmpty bool) (string, error) {
	return NormalizeHTTPURL(rawURL, HTTPURLOptions{
		AllowEmpty:        allowEmpty,
		AllowPrivateHosts: true,
	})
}

// NormalizePublicCallbackURL normalizes an externally supplied callback URL.
func NormalizePublicCallbackURL(rawURL string, allowEmpty bool) (string, error) {
	return NormalizeHTTPURL(rawURL, HTTPURLOptions{
		AllowEmpty: allowEmpty,
	})
}

// NewPublicHTTPClient returns an HTTP client for externally supplied URLs.
func NewPublicHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = publicDialContext

	return &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		CheckRedirect: NoRedirectPolicy,
	}
}

// NoRedirectPolicy keeps signed callback requests on the configured target URL.
func NoRedirectPolicy(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func publicDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	host = strings.Trim(host, "[]")
	if isPrivateOrLocalHost(host) {
		return nil, fmt.Errorf("target host must not point to localhost or private network")
	}

	dialer := &net.Dialer{}
	if ip := net.ParseIP(host); ip != nil {
		return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("target host has no resolved addresses")
	}
	for _, addr := range addrs {
		if isPrivateOrLocalIP(addr.IP) {
			return nil, fmt.Errorf("target host resolved to localhost or private network")
		}
	}
	for _, addr := range addrs {
		if network == "tcp4" && addr.IP.To4() == nil {
			continue
		}
		if network == "tcp6" && addr.IP.To4() != nil {
			continue
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(addr.IP.String(), port))
	}

	return nil, fmt.Errorf("target host has no address for network %s", network)
}

func isPrivateOrLocalHost(host string) bool {
	normalized := strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if normalized == "localhost" || strings.HasSuffix(normalized, ".localhost") {
		return true
	}

	ip := net.ParseIP(normalized)
	if ip == nil {
		return false
	}
	return isPrivateOrLocalIP(ip)
}

func isPrivateOrLocalIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}
