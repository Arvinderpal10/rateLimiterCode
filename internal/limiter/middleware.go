package limiter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"
)

type ctxKey string

const RequestIDKey ctxKey = "requestID"

func WithRequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateSecureRandomID()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next(w, r.WithContext(ctx))
	}
}

func generateSecureRandomID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405") + "-fallback"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID, _ := r.Context().Value(RequestIDKey).(string)
		log.Printf("→ %s %s [reqID=%s]", r.Method, r.URL.Path, reqID)
		next(w, r)
		log.Printf("← %s %s [reqID=%s] %v", r.Method, r.URL.Path, reqID, time.Since(start))
	}
}
