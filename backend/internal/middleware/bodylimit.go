package middleware

import (
	"net/http"
)

const maxBodySize = 1 << 20 // 1MB

// BodyLimit limits request body size to prevent DoS attacks
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		next.ServeHTTP(w, r)
	})
}

