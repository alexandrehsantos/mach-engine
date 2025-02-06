package matching

import (
	"fmt"
	"sync"

	"company.com/matchengine/internal/domain/order"
	"company.com/matchengine/internal/domain/orderbook"
)

type Service struct {
	books map[string]*orderbook.OrderBook
	mutex sync.RWMutex
}

func NewService() *Service {
	return &Service{
		books: make(map[string]*orderbook.OrderBook),
	}
}

func (s *Service) AddOrder(o *order.Order) error {
	s.mutex.Lock()
	book, exists := s.books[o.Symbol]
	if !exists {
		book = orderbook.NewOrderBook(o.Symbol)
		s.books[o.Symbol] = book
	}
	s.mutex.Unlock()

	return book.AddOrder(o)
}

func (s *Service) CancelOrder(symbol, orderID string) error {
	s.mutex.RLock()
	book, exists := s.books[symbol]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("symbol not found: %s", symbol)
	}

	return book.CancelOrder(orderID)
}

func (s *Service) GetOrderBook(symbol string) (*orderbook.OrderBookSnapshot, error) {
	s.mutex.RLock()
	book, exists := s.books[symbol]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	return book.GetOrderBook(), nil
}
