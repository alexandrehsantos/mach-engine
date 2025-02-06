package orderbook

import (
	"company.com/matchengine/internal/domain/order"
)

// Engine define a interface do motor de matching
type Engine interface {
	// Order management
	AddOrder(order *order.Order) error
	CancelOrder(orderID string) error

	// Query methods
	GetOrder(orderID string) (*order.Order, error)
	GetOrderBook(symbol string) (*OrderBookSnapshot, error)

	// Market data
	GetBestBid(symbol string) (price, quantity float64, err error)
	GetBestAsk(symbol string) (price, quantity float64, err error)
}

// OrderBookSnapshot representa um snapshot do order book
type OrderBookSnapshot struct {
	Symbol string       `json:"symbol"`
	Bids   []PriceLevel `json:"bids"`
	Asks   []PriceLevel `json:"asks"`
}
