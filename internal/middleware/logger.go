package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type customLoggerWriter struct {
	w          http.ResponseWriter
	statusCode int
}

func (c *customLoggerWriter) Flush() {
	if flusher, ok := c.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (c *customLoggerWriter) Header() http.Header {
	return c.w.Header()
}

func (c *customLoggerWriter) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *customLoggerWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.w.WriteHeader(statusCode)
}

func Logger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/api/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		cw := &customLoggerWriter{w: w}

		next.ServeHTTP(cw, r)

		log.Info("Request",
			"status", cw.statusCode,
			"method", r.Method,
			"uri", r.RequestURI,
			"remoteAddr", r.RemoteAddr,
			"userAgent", r.UserAgent(),
			"duration", time.Since(start),
		)
	})
}
