package httpio

import (
	"errors"
	"strings"
	"testing"
)

func TestReadAllLimitedRejectsOversizedBody(t *testing.T) {
	_, err := ReadAllLimited(strings.NewReader("abcdef"), 4)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestReadAllLimitedReadsWithinLimit(t *testing.T) {
	body, err := ReadAllLimited(strings.NewReader("abcd"), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "abcd" {
		t.Fatalf("body want abcd got %s", string(body))
	}
}

func TestSnippetTruncatesLongText(t *testing.T) {
	got := Snippet([]byte("  abcdef  "), 3)
	if got != "abc...(truncated)" {
		t.Fatalf("snippet got %q", got)
	}
}
