package order

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Side represents the order side (buy/sell)
type Side string

// Status represents the order status
type Status string

// Constants for order sides
const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

// Constants for order statuses
const (
	StatusNew       Status = "new"
	StatusFilled    Status = "filled"
	StatusCancelled Status = "cancelled"
	StatusPartial   Status = "partial"
)

// Order represents a trading order
type Order struct {
	ID        string    `json:"id"`
	Side      Side      `json:"side"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Filled    float64   `json:"filled"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewOrder creates a new order instance
func NewOrder(side Side, symbol string, price, quantity float64) (*Order, error) {
	if price <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	now := time.Now()
	return &Order{
		ID:        generateOrderID(),
		Side:      side,
		Symbol:    symbol,
		Price:     price,
		Quantity:  quantity,
		Filled:    0,
		Status:    StatusNew,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Fill updates the order's filled quantity and status
func (o *Order) Fill(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("fill quantity must be positive")
	}
	if o.Status == StatusCancelled {
		return fmt.Errorf("cannot fill cancelled order")
	}

	o.Filled += quantity
	o.UpdatedAt = time.Now()

	if o.Filled > o.Quantity {
		return fmt.Errorf("fill amount exceeds order quantity")
	}

	if o.Filled == o.Quantity {
		o.Status = StatusFilled
	} else {
		o.Status = StatusPartial
	}
	return nil
}

// Cancel marks the order as cancelled
func (o *Order) Cancel() error {
	if o.Status == StatusFilled {
		return fmt.Errorf("cannot cancel filled order")
	}
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

// RemainingQuantity returns the unfilled quantity
func (o *Order) RemainingQuantity() float64 {
	return o.Quantity - o.Filled
}

// IsActive returns whether the order is still active
func (o *Order) IsActive() bool {
	return o.Status != StatusFilled && o.Status != StatusCancelled
}

func generateOrderID() string {
	return uuid.New().String()
}
