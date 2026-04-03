package middleware

import (
	"context"
	"crypto/rand"
	"math"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDLength = 12
)

type requestIDContextKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := normalizeRequestID(r.Header.Get(requestIDHeader))
		if reqID == "" {
			reqID = newRequestID()
		}

		ctx := context.WithValue(r.Context(), requestIDContextKey{}, reqID)
		w.Header().Set(requestIDHeader, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRequestID(ctx context.Context) string {
	if value := ctx.Value(requestIDContextKey{}); value != nil {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func normalizeRequestID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len(trimmed) != requestIDLength {
		return ""
	}
	upper := strings.ToUpper(trimmed)
	for i := 0; i < len(upper); i++ {
		ch := upper[i]
		if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') {
			continue
		}
		return ""
	}
	return upper
}

func newRequestID() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, requestIDLength)
	if _, err := rand.Read(buf); err == nil {
		out := make([]byte, len(buf))
		for i, b := range buf {
			out[i] = alphabet[int(b)%len(alphabet)]
		}
		return string(out)
	}

	seed := time.Now().UnixNano()
	if seed == 0 {
		seed = math.MaxInt64 - 1
	}
	rng := mathrand.New(mathrand.NewSource(seed))
	out := make([]byte, requestIDLength)
	for i := 0; i < requestIDLength; i++ {
		out[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(out)
}
