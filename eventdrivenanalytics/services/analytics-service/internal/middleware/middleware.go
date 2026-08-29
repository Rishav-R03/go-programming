package middleware

import (
	"net/http"
	"strconv"
	"time"

	"analyticsservice/internal/metrics"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			path := r.URL.Path
			method := r.Method
			status := strconv.Itoa(rw.statusCode)

			m.HTTPRequestDurationSeconds.WithLabelValues(method, path).Observe(duration)
			m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		})
	}
}
