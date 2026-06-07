package urlguard

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizePublicCallbackURLRejectsLocalTargets(t *testing.T) {
	cases := []string{
		"http://localhost/callback",
		"http://127.0.0.1/callback",
		"http://10.0.0.1/callback",
		"ftp://example.com/callback",
		"https://user:pass@example.com/callback",
	}
	for _, rawURL := range cases {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := NormalizePublicCallbackURL(rawURL, false); err == nil {
				t.Fatalf("expected %q to be rejected", rawURL)
			}
		})
	}
}

func TestNormalizeTrustedCallbackURLAllowsPrivateHosts(t *testing.T) {
	got, err := NormalizeTrustedCallbackURL(" HTTP://127.0.0.1:8444/internal/order-fulfilled#frag ", false)
	if err != nil {
		t.Fatalf("trusted callback should allow private hosts: %v", err)
	}
	if got != "http://127.0.0.1:8444/internal/order-fulfilled" {
		t.Fatalf("unexpected normalized url: %s", got)
	}
}

func TestNewPublicHTTPClientRejectsResolvedPrivateAddress(t *testing.T) {
	client := NewPublicHTTPClient(100 * time.Millisecond)
	_, err := client.Get("http://localhost/callback")
	if err == nil {
		t.Fatal("expected localhost request to be rejected")
	}
	if !strings.Contains(err.Error(), "localhost or private network") {
		t.Fatalf("unexpected error: %v", err)
	}
}
