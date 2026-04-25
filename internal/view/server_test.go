package view

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestNotFound(t *testing.T) {
	srv := New(WithLogger(zerolog.New(os.Stdout)))
	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	rec := httptest.NewRecorder()
	srv.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
