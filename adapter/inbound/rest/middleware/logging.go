package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	ansiReset = "\x1b[0m"
)

// Logging logs inbound HTTP requests and their responses.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		reqID := getRequestID(r.Context())
		fmt.Fprintf(
			os.Stdout,
			"%s ▸ %s %s %s %s\n",
			formatLogTime(start),
			colorizeLabel("REQ", 6),
			reqID,
			colorizeMethod(r.Method),
			path,
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		statusText := colorizeStatus(status)
		duration := time.Since(start)
		fmt.Fprintf(
			os.Stdout,
			"%s ▸ %s %s %s %s\n",
			formatLogTime(time.Now()),
			colorizeLabel("RES", 13),
			reqID,
			statusText,
			duration,
		)
	})
}

func colorizeStatus(status int) string {
	text := fmt.Sprintf("%d", status)
	if status >= 500 {
		return colorizeStatusBackground(text, 9)
	}
	if status >= 300 {
		return colorizeStatusBackground(text, 11)
	}
	if status >= 200 {
		return colorizeStatusBackground(text, 10)
	}
	return colorizeStatusBackground(text, 8)
}

func formatLogTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

func colorizeLabel(label string, color int) string {
	return fmt.Sprintf("%s%s%s", ansiFg256(color), label, ansiReset)
}

func colorizeMethod(method string) string {
	switch method {
	case http.MethodGet:
		return colorizeMethodBackground(method, 4)
	case http.MethodPost:
		return colorizeMethodBackground(method, 2)
	case http.MethodPut:
		return colorizeMethodBackground(method, 3)
	case http.MethodPatch:
		return colorizeMethodBackground(method, 5)
	case http.MethodDelete:
		return colorizeMethodBackground(method, 1)
	default:
		return colorizeMethodBackground(method, 8)
	}
}

func colorizeStatusBackground(text string, bgColor int) string {
	return fmt.Sprintf("%s%s%s%s", ansiBg256(bgColor), ansiFg256(0), text, ansiReset)
}

func colorizeMethodBackground(text string, bgColor int) string {
	return fmt.Sprintf("%s%s%s%s", ansiBg256(bgColor), ansiFg256(0), text, ansiReset)
}

func ansiFg256(color int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", color)
}

func ansiBg256(color int) string {
	return fmt.Sprintf("\x1b[48;5;%dm", color)
}
