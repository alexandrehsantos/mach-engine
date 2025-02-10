package matching

import (
	"fmt"
	"log"
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
	defer s.mutex.Unlock()

	book, exists := s.books[o.Symbol]
	if !exists {
		book = orderbook.NewOrderBook(o.Symbol)
		s.books[o.Symbol] = book
	}
	log.Printf("Adding order %s to book %s", o.ID, o.Symbol)
	return book.AddOrder(o)
}

func (s *Service) CancelOrder(symbol, orderID string) error {
	s.mutex.RLock()
	book, exists := s.books[symbol]
	s.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("symbol not found: %s", symbol)
	}
	log.Printf("Canceling order %s from book %s", orderID, symbol)
	return book.CancelOrder(orderID)
}

func (s *Service) GetOrder(orderID string) (*order.Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	log.Printf("Looking for order %s in %d books", orderID, len(s.books))
	// Procura a ordem em todos os order books
	for symbol, book := range s.books {
		log.Printf("Checking book %s", symbol)
		if o, err := book.GetOrder(orderID); err == nil {
			return o, nil
		}
	}

	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (s *Service) GetOrderBook(symbol string) (*orderbook.OrderBookSnapshot, error) {
	s.mutex.RLock()
	book, exists := s.books[symbol]
	s.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}
	log.Printf("Getting order book for symbol %s", symbol)
	return book.GetOrderBook(), nil
}
