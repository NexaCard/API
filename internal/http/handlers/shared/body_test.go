package shared

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReadRequestBodyWithLimitRestoresBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("hello"))

	body, err := ReadRequestBodyWithLimit(c, 16)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("unexpected body: %q", string(body))
	}
	again, err := io.ReadAll(c.Request.Body)
	if err != nil {
		t.Fatalf("read restored body failed: %v", err)
	}
	if string(again) != "hello" {
		t.Fatalf("unexpected restored body: %q", string(again))
	}
}

func TestReadRequestBodyWithLimitRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("oversized"))

	_, err := ReadRequestBodyWithLimit(c, 4)
	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("expected ErrRequestBodyTooLarge, got %v", err)
	}
}
