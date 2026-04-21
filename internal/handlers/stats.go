package handlers

import (
	"net/http"

	"github.com/arvinderpal10/ratelimiter/internal/limiter"
	"github.com/arvinderpal10/ratelimiter/pkg/response"
)

type StatsHandler struct {
	Store *limiter.Store
}

func NewStatsHandler(store *limiter.Store) *StatsHandler {
	return &StatsHandler{Store: store}
}

func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.Store.Stats()
	response.JSON(w, http.StatusOK, map[string]interface{}{"users": stats})
}
