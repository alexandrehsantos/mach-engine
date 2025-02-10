package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"company.com/matchengine/internal/domain/order"
	"company.com/matchengine/internal/service/matching"
)

type Handler struct {
	service *matching.Service
	mux     *http.ServeMux
}

func NewHandler(service *matching.Service) *Handler {
	h := &Handler{
		service: service,
		mux:     http.NewServeMux(),
	}

	// Registra handlers
	h.mux.HandleFunc("/orders/", h.handleOrder)    // Precisa vir primeiro
	h.mux.HandleFunc("/orders", h.handleOrders)    // Depois o mais específico
	h.mux.HandleFunc("/health", HealthCheck)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	h.mux.ServeHTTP(w, r)
}

type OrderRequest struct {
	Symbol   string     `json:"symbol"`
	Side     order.Side `json:"side"`
	Price    float64    `json:"price"`
	Quantity float64    `json:"quantity"`
}

func (h *Handler) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	o, err := order.NewOrder(req.Side, req.Symbol, req.Price, req.Quantity)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := h.service.AddOrder(o); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

func (h *Handler) handleOrder(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleOrder: path=%s", r.URL.Path)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	log.Printf("handleOrder: parts=%v", parts)
	if len(parts) != 2 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("invalid path: %s", r.URL.Path)})
		return
	}

	orderID := parts[1]
	if orderID == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "order ID is required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getOrder(w, r, orderID)
	case http.MethodDelete:
		h.cancelOrder(w, r, orderID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	o, err := h.service.GetOrder(orderID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(o)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	o, err := h.service.GetOrder(orderID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := h.service.CancelOrder(o.Symbol, orderID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
}
