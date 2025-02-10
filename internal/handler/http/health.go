package http

import (
	"encoding/json"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Define o header Content-Type para todas as respostas
	w.Header().Set("Content-Type", "application/json")

	// Se houver um request ID, inclui no response
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		w.Header().Set("X-Request-ID", reqID)
	}

	// Verifica se o método é GET
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "method not allowed",
		})
		return
	}

	// Retorna o status
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
