package orderbook

import (
	"testing"

	"company.com/matchengine/internal/domain/order"
)

func TestOrderBook_AddOrder(t *testing.T) {
	ob := NewOrderBook("BTC-USD")

	tests := []struct {
		name    string
		order   *order.Order
		wantErr bool
	}{
		{
			name: "valid buy order",
			order: order.NewOrder(
				order.Buy,
				"BTC-USD",
				50000.0,
				1.0,
			),
			wantErr: false,
		},
		{
			name: "valid sell order",
			order: order.NewOrder(
				order.Sell,
				"BTC-USD",
				50100.0,
				1.0,
			),
			wantErr: false,
		},
		{
			name: "invalid symbol",
			order: order.NewOrder(
				order.Buy,
				"ETH-USD",
				50000.0,
				1.0,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ob.AddOrder(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderBook.AddOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderBook_Match(t *testing.T) {
	ob := NewOrderBook("BTC-USD")

	// Adiciona ordem de compra
	buyOrder := order.NewOrder(
		order.Buy,
		"BTC-USD",
		50000.0,
		2.0,
	)
	ob.AddOrder(buyOrder)

	// Adiciona ordem de venda que deve casar parcialmente
	sellOrder := order.NewOrder(
		order.Sell,
		"BTC-USD",
		50000.0,
		1.0,
	)
	ob.AddOrder(sellOrder)

	// Verifica se o matching ocorreu corretamente
	if buyOrder.Status != order.Partial {
		t.Errorf("expected buy order status to be %v, got %v", order.Partial, buyOrder.Status)
	}
	if sellOrder.Status != order.Filled {
		t.Errorf("expected sell order status to be %v, got %v", order.Filled, sellOrder.Status)
	}
	if buyOrder.Filled != 1.0 {
		t.Errorf("expected buy order filled quantity to be 1.0, got %v", buyOrder.Filled)
	}
	if sellOrder.Filled != 1.0 {
		t.Errorf("expected sell order filled quantity to be 1.0, got %v", sellOrder.Filled)
	}
}

func TestOrderBook_CancelOrder(t *testing.T) {
	ob := NewOrderBook("BTC-USD")

	// Adiciona ordem de compra
	order := order.NewOrder(
		order.Buy,
		"BTC-USD",
		50000.0,
		1.0,
	)
	ob.AddOrder(order)

	// Tenta cancelar
	err := ob.CancelOrder(order.ID)
	if err != nil {
		t.Errorf("unexpected error canceling order: %v", err)
	}

	// Verifica se a ordem foi cancelada
	if order.Status != order.Cancelled {
		t.Errorf("expected order status to be %v, got %v", order.Cancelled, order.Status)
	}

	// Tenta cancelar ordem inexistente
	err = ob.CancelOrder("invalid-id")
	if err == nil {
		t.Error("expected error canceling invalid order")
	}
}

func TestOrderBook_GetBestPrices(t *testing.T) {
	ob := NewOrderBook("BTC-USD")

	// Inicialmente não deve ter preços
	_, _, err := ob.GetBestBid()
	if err == nil {
		t.Error("expected error when no bids available")
	}

	_, _, err = ob.GetBestAsk()
	if err == nil {
		t.Error("expected error when no asks available")
	}

	// Adiciona ordens
	buyOrder := order.NewOrder(
		order.Buy,
		"BTC-USD",
		50000.0,
		1.0,
	)
	ob.AddOrder(buyOrder)

	sellOrder := order.NewOrder(
		order.Sell,
		"BTC-USD",
		50100.0,
		1.0,
	)
	ob.AddOrder(sellOrder)

	// Verifica melhor bid
	price, qty, err := ob.GetBestBid()
	if err != nil {
		t.Errorf("unexpected error getting best bid: %v", err)
	}
	if price != 50000.0 {
		t.Errorf("expected best bid price to be 50000.0, got %v", price)
	}
	if qty != 1.0 {
		t.Errorf("expected best bid quantity to be 1.0, got %v", qty)
	}

	// Verifica melhor ask
	price, qty, err = ob.GetBestAsk()
	if err != nil {
		t.Errorf("unexpected error getting best ask: %v", err)
	}
	if price != 50100.0 {
		t.Errorf("expected best ask price to be 50100.0, got %v", price)
	}
	if qty != 1.0 {
		t.Errorf("expected best ask quantity to be 1.0, got %v", qty)
	}
}
