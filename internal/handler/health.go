package handler

import (
	"encoding/json"
	"net/http"
	"vturb-go/internal/config"
)

func HealthCheck(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "ok",
			"env":    cfg.Env,
			"port":   cfg.Port,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
