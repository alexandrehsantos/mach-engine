package errors

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, data interface{}) {
	var resp Response
	switch v := data.(type) {
	case *APIError:
		resp = Response{
			Success: false,
			Error:   v,
		}
		w.WriteHeader(v.Status)
	default:
		resp = Response{
			Success: true,
			Data:    data,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
