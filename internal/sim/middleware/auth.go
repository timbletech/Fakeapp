package middleware

import (
	"encoding/json"
	"net/http"

	"device_only/internal/sim/model"

	"github.com/google/uuid"
)

// APIKeyAuth returns an http.Handler that rejects requests whose X-API-Key header
// does not exactly match apiKey. All /v1/auth/* routes must be wrapped with this.
func APIKeyAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" || key != apiKey {
			requestID := "req_" + uuid.New().String()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error:     "unauthorized",
				Message:   "Invalid or missing API key.",
				RequestID: requestID,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
