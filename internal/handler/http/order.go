package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"company.com/matchengine/internal/domain/order"
	"company.com/matchengine/internal/service/matching"
	"company.com/matchengine/pkg/errors"
)

type OrderHandler struct {
	matchingService *matching.Service
}

func NewOrderHandler(service *matching.Service) *OrderHandler {
	if service == nil {
		panic("matching service cannot be nil")
	}
	return &OrderHandler{
		matchingService: service,
	}
}

type CreateOrderRequest struct {
	Symbol   string     `json:"symbol" validate:"required"`
	Side     order.Side `json:"side" validate:"required,oneof=buy sell"`
	Price    float64    `json:"price" validate:"required,gt=0"`
	Quantity float64    `json:"quantity" validate:"required,gt=0"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.NewBadRequest("Invalid request body"))
		return
	}
	defer r.Body.Close()

	// Validate request
	if err := validate(req); err != nil {
		errors.WriteJSON(w, errors.NewBadRequest(err.Error()))
		return
	}

	// Create order
	newOrder, err := order.NewOrder(req.Side, req.Symbol, req.Price, req.Quantity)
	if err != nil {
		errors.WriteJSON(w, errors.NewBadRequest(err.Error()))
		return
	}

	// Add to matching engine
	if err := h.matchingService.AddOrder(newOrder); err != nil {
		errors.WriteJSON(w, errors.NewInternal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}

func validate(req CreateOrderRequest) error {
	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return nil
}
