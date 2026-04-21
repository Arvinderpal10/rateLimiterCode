package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arvinderpal10/ratelimiter/internal/limiter"
	"github.com/arvinderpal10/ratelimiter/pkg/response"
)

type RequestPayload struct {
	UserID  string `json:"user_id"`
	Payload string `json:"payload"`
}

type RequestHandler struct {
	Store *limiter.Store
}

func NewRequestHandler(store *limiter.Store) *RequestHandler {
	return &RequestHandler{Store: store}
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		response.JSONError(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if !h.Store.Allow(req.UserID) {
		w.Header().Set("Retry-After", "60")
		response.JSONError(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	reqID, _ := r.Context().Value(limiter.RequestIDKey).(string)
	log.Printf("[reqID=%s] processed request for user %s: %s", reqID, req.UserID, req.Payload)

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
