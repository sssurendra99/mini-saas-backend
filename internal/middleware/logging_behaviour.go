package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// I'm using the Response writer to capture the status code. so that I'm creating a wrapper
type ResponseWriterWrapper struct {
	http.ResponseWriter
	status int
}

// We do this because once the status code is sent we can not get it so we try to store it before sending it.
func (rww *ResponseWriterWrapper) WriteHeader(status int) {
	rww.status = status
	rww.ResponseWriter.WriteHeader(status)
}

func LoggingBehavior(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		starting_time := time.Now()

		rec := &ResponseWriterWrapper{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(w, r)

		slog.Info(
			"Request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(starting_time)),
		)
	})
}
