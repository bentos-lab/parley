package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRequestLoggingMiddleware_FormatsAndRequestID(t *testing.T) {
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Use(Logging)
	router.Get("/api/debates", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var buf bytes.Buffer
	stdout := os.Stdout
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = wPipe
	t.Cleanup(func() {
		os.Stdout = stdout
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	_ = wPipe.Close()
	_, _ = buf.ReadFrom(rPipe)

	raw := buf.String()
	stripped := stripANSI(raw)

	if !strings.Contains(stripped, " REQ ") || !strings.Contains(stripped, " RES ") {
		t.Fatalf("expected REQ and RES lines, got: %q", stripped)
	}
	if !strings.Contains(stripped, " ▸ ") {
		t.Fatalf("expected timestamp separator, got: %q", stripped)
	}
	if strings.Contains(stripped, "WARNING") || strings.Contains(stripped, "ERROR") {
		t.Fatalf("unexpected WARNING/ERROR lines, got: %q", stripped)
	}

	lines := strings.Split(strings.TrimSpace(stripped), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 log lines, got: %q", stripped)
	}

	inFields := strings.Fields(lines[0])
	outFields := strings.Fields(lines[1])
	if len(inFields) < 6 || len(outFields) < 6 {
		t.Fatalf("unexpected log format, in: %q out: %q", lines[0], lines[1])
	}

	inReqID := fieldAfter(inFields, "REQ")
	outReqID := fieldAfter(outFields, "RES")
	if inReqID == "" || inReqID != outReqID {
		t.Fatalf("request id mismatch, in: %q out: %q", inReqID, outReqID)
	}

	backgroundColor := regexp.MustCompile("\x1b\\[48;")
	if !backgroundColor.MatchString(raw) {
		t.Fatalf("expected status to use background color, got: %q", raw)
	}

	methodColor := regexp.MustCompile("\x1b\\[48;5;[0-9]+m\x1b\\[38;5;0mGET")
	if !methodColor.MatchString(raw) {
		t.Fatalf("expected method to be colorized, got: %q", raw)
	}
}

func stripANSI(s string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*m")
	return re.ReplaceAllString(s, "")
}

func fieldAfter(fields []string, label string) string {
	for i, value := range fields {
		if value == label && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}
