package http

import (
	"net/http"

	"company.com/matchengine/pkg/errors"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	errors.WriteJSON(w, map[string]string{"status": "ok"})
}
