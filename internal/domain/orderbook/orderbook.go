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

	// Adiciona a ordem ao mapa primeiro
	ob.orders[o.ID] = o

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
	}

	// Limpa níveis vazios após o matching
	ob.cleanupEmptyLevels()

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

	o, exists := ob.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return o, nil
}

// CancelOrder cancela uma ordem existente
func (ob *OrderBook) CancelOrder(orderID string) error {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	o, exists := ob.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Marca a ordem como cancelada
	o.Status = order.StatusCancelled

	// Remove a ordem do nível de preço
	switch o.Side {
	case order.SideBuy:
		ob.removeBuyOrder(o)
	case order.SideSell:
		ob.removeSellOrder(o)
	}

	// Limpa níveis vazios após a remoção
	ob.cleanupEmptyLevels()

	return nil
}

func (ob *OrderBook) removeBuyOrder(o *order.Order) {
	current := ob.buyLevels
	for current != nil {
		if current.Price == o.Price {
			for i, order := range current.Orders {
				if order.ID == o.ID {
					current.Orders = append(current.Orders[:i], current.Orders[i+1:]...)
					break
				}
			}
			break
		}
		current = current.Next
	}
}

func (ob *OrderBook) removeSellOrder(o *order.Order) {
	current := ob.sellLevels
	for current != nil {
		if current.Price == o.Price {
			for i, order := range current.Orders {
				if order.ID == o.ID {
					current.Orders = append(current.Orders[:i], current.Orders[i+1:]...)
					break
				}
			}
			break
		}
		current = current.Next
	}
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
	switch o.Side {
	case order.SideBuy:
		// Se é uma ordem de compra, procura ordens de venda compatíveis
		for ob.sellLevels != nil && o.RemainingQuantity() > 0 {
			sellLevel := ob.sellLevels
			if sellLevel.Price > o.Price {
				// Não há mais ordens compatíveis
				break
			}

			// Processa ordens neste nível de preço
			for i := 0; i < len(sellLevel.Orders) && o.RemainingQuantity() > 0; i++ {
				sellOrder := sellLevel.Orders[i]
				matchQty := min(o.RemainingQuantity(), sellOrder.RemainingQuantity())

				// Executa o match
				o.Fill(matchQty)
				sellOrder.Fill(matchQty)

				// Se a ordem de venda foi totalmente preenchida, remove do nível
				if sellOrder.Status == order.StatusFilled {
					sellLevel.Orders = append(sellLevel.Orders[:i], sellLevel.Orders[i+1:]...)
					i--
				}
			}

			// Se o nível está vazio, remove-o
			if len(sellLevel.Orders) == 0 {
				ob.sellLevels = ob.sellLevels.Next
			}
		}

	case order.SideSell:
		// Se é uma ordem de venda, procura ordens de compra compatíveis
		for ob.buyLevels != nil && o.RemainingQuantity() > 0 {
			buyLevel := ob.buyLevels
			if buyLevel.Price < o.Price {
				// Não há mais ordens compatíveis
				break
			}

			// Processa ordens neste nível de preço
			for i := 0; i < len(buyLevel.Orders) && o.RemainingQuantity() > 0; i++ {
				buyOrder := buyLevel.Orders[i]
				matchQty := min(o.RemainingQuantity(), buyOrder.RemainingQuantity())

				// Executa o match
				o.Fill(matchQty)
				buyOrder.Fill(matchQty)

				// Se a ordem de compra foi totalmente preenchida, remove do nível
				if buyOrder.Status == order.StatusFilled {
					buyLevel.Orders = append(buyLevel.Orders[:i], buyLevel.Orders[i+1:]...)
					i--
				}
			}

			// Se o nível está vazio, remove-o
			if len(buyLevel.Orders) == 0 {
				ob.buyLevels = ob.buyLevels.Next
			}
		}
	}

	return nil
}
