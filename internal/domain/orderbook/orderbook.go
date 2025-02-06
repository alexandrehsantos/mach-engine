package orderbook

import (
	"fmt"
	"sync"

	"company.com/matchengine/internal/domain/order"
)

// PriceLevel representa um nível de preço no order book
type PriceLevel struct {
	Price    float64
	Orders   []*order.Order
	Next     *PriceLevel
	Previous *PriceLevel
}

// OrderBook representa o livro de ordens usando uma lista duplamente encadeada
type OrderBook struct {
	symbol     string
	buyLevels  *PriceLevel
	sellLevels *PriceLevel
	orders     map[string]*order.Order
	mutex      sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		symbol: symbol,
		orders: make(map[string]*order.Order),
	}
}

// AddOrder adiciona uma ordem ao livro
func (ob *OrderBook) AddOrder(o *order.Order) error {
	if o.Symbol != ob.symbol {
		return fmt.Errorf("invalid symbol: %s", o.Symbol)
	}

	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	// Try to match the order first
	if err := ob.tryMatch(o); err != nil {
		return err
	}

	// If order is not fully filled, add to book
	if o.Status != order.StatusFilled {
		switch o.Side {
		case order.SideBuy:
			ob.addBuyOrder(o)
		case order.SideSell:
			ob.addSellOrder(o)
		}
		ob.orders[o.ID] = o
	}

	// Process the match after adding the order
	ob.match()

	return nil
}

func (ob *OrderBook) addBuyOrder(o *order.Order) {
	level := ob.findOrCreateBuyLevel(o.Price)
	level.Orders = append(level.Orders, o)
}

func (ob *OrderBook) addSellOrder(o *order.Order) {
	level := ob.findOrCreateSellLevel(o.Price)
	level.Orders = append(level.Orders, o)
}

// findOrCreateBuyLevel encontra ou cria um nível de preço de compra
func (ob *OrderBook) findOrCreateBuyLevel(price float64) *PriceLevel {
	if ob.buyLevels == nil || price > ob.buyLevels.Price {
		ob.buyLevels = &PriceLevel{
			Price: price,
			Next:  ob.buyLevels,
		}
		return ob.buyLevels
	}

	current := ob.buyLevels
	for current.Next != nil && price < current.Next.Price {
		current = current.Next
	}

	if current.Price == price {
		return current
	}

	newLevel := &PriceLevel{
		Price: price,
		Next:  current.Next,
	}
	current.Next = newLevel
	return newLevel
}

// findOrCreateSellLevel encontra ou cria um nível de preço de venda
func (ob *OrderBook) findOrCreateSellLevel(price float64) *PriceLevel {
	if ob.sellLevels == nil || price < ob.sellLevels.Price {
		ob.sellLevels = &PriceLevel{
			Price: price,
			Next:  ob.sellLevels,
		}
		return ob.sellLevels
	}

	current := ob.sellLevels
	for current.Next != nil && price > current.Next.Price {
		current = current.Next
	}

	if current.Price == price {
		return current
	}

	newLevel := &PriceLevel{
		Price: price,
		Next:  current.Next,
	}
	current.Next = newLevel
	return newLevel
}

// match tenta casar ordens compatíveis
func (ob *OrderBook) match() {
	for ob.buyLevels != nil && ob.sellLevels != nil {
		bestBuy := ob.buyLevels
		bestSell := ob.sellLevels

		// Se não há match possível, para
		if bestBuy.Price < bestSell.Price {
			break
		}

		// Processa ordens neste nível de preço
		ob.processLevelMatch(bestBuy, bestSell)

		// Remove níveis vazios
		ob.cleanupEmptyLevels()
	}
}

func (ob *OrderBook) processLevelMatch(buyLevel, sellLevel *PriceLevel) {
	for len(buyLevel.Orders) > 0 && len(sellLevel.Orders) > 0 {
		buy := buyLevel.Orders[0]
		sell := sellLevel.Orders[0]

		// Calculate match quantity
		matchQty := min(buy.RemainingQuantity(), sell.RemainingQuantity())

		// Execute the match
		buy.Fill(matchQty)
		sell.Fill(matchQty)

		// Remove filled orders
		if buy.Status == order.StatusFilled {
			buyLevel.Orders = buyLevel.Orders[1:]
		}
		if sell.Status == order.StatusFilled {
			sellLevel.Orders = sellLevel.Orders[1:]
		}
	}
}

func (ob *OrderBook) cleanupEmptyLevels() {
	// Limpa níveis vazios de compra
	for ob.buyLevels != nil && len(ob.buyLevels.Orders) == 0 {
		ob.buyLevels = ob.buyLevels.Next
		if ob.buyLevels != nil {
			ob.buyLevels.Previous = nil
		}
	}

	// Limpa níveis vazios de venda
	for ob.sellLevels != nil && len(ob.sellLevels.Orders) == 0 {
		ob.sellLevels = ob.sellLevels.Next
		if ob.sellLevels != nil {
			ob.sellLevels.Previous = nil
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GetOrder retorna uma ordem pelo ID
func (ob *OrderBook) GetOrder(orderID string) (*order.Order, error) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	// Procura nas ordens de compra
	if order := ob.findOrder(ob.buyLevels, orderID); order != nil {
		return order, nil
	}

	// Procura nas ordens de venda
	if order := ob.findOrder(ob.sellLevels, orderID); order != nil {
		return order, nil
	}

	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (ob *OrderBook) findOrder(level *PriceLevel, orderID string) *order.Order {
	for ; level != nil; level = level.Next {
		for _, o := range level.Orders {
			if o.ID == orderID {
				return o
			}
		}
	}
	return nil
}

// CancelOrder cancela uma ordem existente
func (ob *OrderBook) CancelOrder(orderID string) error {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	o, exists := ob.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if err := o.Cancel(); err != nil {
		return err
	}

	delete(ob.orders, orderID)
	return nil
}

// GetOrderBook retorna um snapshot do order book
func (ob *OrderBook) GetOrderBook() *OrderBookSnapshot {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	snapshot := &OrderBookSnapshot{
		Symbol: ob.symbol,
		Bids:   make([]PriceLevel, 0),
		Asks:   make([]PriceLevel, 0),
	}

	// Add bids
	for level := ob.buyLevels; level != nil; level = level.Next {
		snapshot.Bids = append(snapshot.Bids, *level)
	}

	// Add asks
	for level := ob.sellLevels; level != nil; level = level.Next {
		snapshot.Asks = append(snapshot.Asks, *level)
	}

	return snapshot
}

// GetBestBid retorna o melhor preço de compra
func (ob *OrderBook) GetBestBid() (price, quantity float64, err error) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if ob.buyLevels == nil || len(ob.buyLevels.Orders) == 0 {
		return 0, 0, fmt.Errorf("no bids available")
	}

	level := ob.buyLevels
	totalQty := 0.0
	for _, o := range level.Orders {
		totalQty += o.RemainingQuantity()
	}

	return level.Price, totalQty, nil
}

// GetBestAsk retorna o melhor preço de venda
func (ob *OrderBook) GetBestAsk() (price, quantity float64, err error) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if ob.sellLevels == nil || len(ob.sellLevels.Orders) == 0 {
		return 0, 0, fmt.Errorf("no asks available")
	}

	level := ob.sellLevels
	totalQty := 0.0
	for _, o := range level.Orders {
		totalQty += o.RemainingQuantity()
	}

	return level.Price, totalQty, nil
}

func (ob *OrderBook) tryMatch(o *order.Order) error {
	var matchingLevels *PriceLevel
	var isAggressive bool

	switch o.Side {
	case order.SideBuy:
		matchingLevels = ob.sellLevels
		isAggressive = true
	case order.SideSell:
		matchingLevels = ob.buyLevels
		isAggressive = false
	}

	for matchingLevels != nil && o.Status != order.StatusFilled {
		if (isAggressive && o.Price < matchingLevels.Price) ||
			(!isAggressive && o.Price > matchingLevels.Price) {
			break
		}

		for _, restingOrder := range matchingLevels.Orders {
			if restingOrder.Status == order.StatusCancelled {
				continue
			}

			matchQty := min(o.RemainingQuantity(), restingOrder.RemainingQuantity())
			if matchQty <= 0 {
				continue
			}

			// Execute the match
			if err := o.Fill(matchQty); err != nil {
				return err
			}
			if err := restingOrder.Fill(matchQty); err != nil {
				return err
			}

			if restingOrder.Status == order.StatusFilled {
				delete(ob.orders, restingOrder.ID)
			}

			if o.Status == order.StatusFilled {
				break
			}
		}

		matchingLevels = matchingLevels.Next
	}

	return nil
}
