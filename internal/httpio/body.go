package httpio

import (
	"errors"
	"io"
	"strings"
)

const (
	DefaultMaxResponseBodyBytes int64 = 1 << 20
	DefaultMaxLogSnippetBytes         = 4096
)

var ErrBodyTooLarge = errors.New("response body too large")

func ReadAllLimited(r io.Reader, limit int64) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = DefaultMaxResponseBodyBytes
	}
	body, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrBodyTooLarge
	}
	return body, nil
}

func Snippet(raw []byte, maxBytes int) string {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxLogSnippetBytes
	}
	text := strings.TrimSpace(string(raw))
	if len(text) <= maxBytes {
		return text
	}
	return text[:maxBytes] + "...(truncated)"
}
