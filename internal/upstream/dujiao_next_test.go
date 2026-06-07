package upstream

import (
	"net/url"
	"testing"

	"github.com/NexaCard/API/internal/models"
)

func TestResolveImageDownloadURLRestrictsToUpstreamHost(t *testing.T) {
	adapter := NewDujiaoNextAdapter(&models.SiteConnection{
		BaseURL:   "https://upstream.example.com",
		ApiKey:    "key",
		ApiSecret: "secret",
	}, t.TempDir())

	got, extSource, err := adapter.resolveImageDownloadURL("/uploads/product.png")
	if err != nil {
		t.Fatalf("root-relative image should be accepted: %v", err)
	}
	if got != "https://upstream.example.com/uploads/product.png" {
		t.Fatalf("unexpected image url: %s", got)
	}
	if extSource != "/uploads/product.png" {
		t.Fatalf("unexpected ext source: %s", extSource)
	}

	if _, _, err := adapter.resolveImageDownloadURL("https://cdn.example.com/product.png"); err == nil {
		t.Fatal("cross-host image should be rejected")
	}
	if _, _, err := adapter.resolveImageDownloadURL("product.png"); err == nil {
		t.Fatal("non-root relative image should be rejected")
	}
}

func TestSameURLAuthorityTreatsDefaultPortsAsEquivalent(t *testing.T) {
	a, _ := url.Parse("https://upstream.example.com")
	b, _ := url.Parse("https://upstream.example.com:443/path")
	c, _ := url.Parse("http://upstream.example.com:80/path")

	if !sameURLAuthority(a, b) {
		t.Fatal("https default port should match explicit 443")
	}
	if sameURLAuthority(a, c) {
		t.Fatal("different schemes should not match")
	}
}
