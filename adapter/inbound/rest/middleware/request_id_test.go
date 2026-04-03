package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRequestIDMiddleware_UsesValidHeader(t *testing.T) {
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Get("/api/debates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-ReqID-Context", getRequestID(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates", nil)
	req.Header.Set(requestIDHeader, "a1b2c3d4e5f6")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got != "A1B2C3D4E5F6" {
		t.Fatalf("expected header to be normalized, got %q", got)
	}
	if got := rec.Header().Get("X-ReqID-Context"); got != "A1B2C3D4E5F6" {
		t.Fatalf("expected context id to match, got %q", got)
	}
}

func TestRequestIDMiddleware_GeneratesWhenInvalidHeader(t *testing.T) {
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Get("/api/debates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-ReqID-Context", getRequestID(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates", nil)
	req.Header.Set(requestIDHeader, "invalid")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	re := regexp.MustCompile("^[A-Z0-9]{12}$")
	headerID := rec.Header().Get(requestIDHeader)
	contextID := rec.Header().Get("X-ReqID-Context")
	if !re.MatchString(headerID) {
		t.Fatalf("expected generated header id to match base36, got %q", headerID)
	}
	if headerID != contextID {
		t.Fatalf("expected context id to match header, got %q vs %q", contextID, headerID)
	}
}
