package shared

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var ErrRequestBodyTooLarge = errors.New("request body too large")

func ReadRequestBodyWithLimit(c *gin.Context, limit int64) ([]byte, error) {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return nil, nil
	}
	if limit > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) || strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			return nil, ErrRequestBodyTooLarge
		}
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
