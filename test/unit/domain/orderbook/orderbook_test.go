package orderbook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"company.com/matchengine/internal/domain/order"
	"company.com/matchengine/internal/domain/orderbook"
)

func TestOrderBook_AddOrder(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		side    order.Side
		price   float64
		qty     float64
		wantErr bool
	}{
		{
			name:    "valid buy order",
			symbol:  "BTC-USD",
			side:    order.SideBuy,
			price:   50000.0,
			qty:     1.0,
			wantErr: false,
		},
		{
			name:    "valid sell order",
			symbol:  "BTC-USD",
			side:    order.SideSell,
			price:   50000.0,
			qty:     1.0,
			wantErr: false,
		},
		{
			name:    "invalid symbol",
			symbol:  "ETH-USD",
			side:    order.SideBuy,
			price:   50000.0,
			qty:     1.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := orderbook.NewOrderBook("BTC-USD")

			o, err := order.NewOrder(tt.side, tt.symbol, tt.price, tt.qty)
			require.NoError(t, err)

			err = book.AddOrder(o)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestOrderBook_Match(t *testing.T) {
	book := orderbook.NewOrderBook("BTC-USD")

	// Create buy order
	buyOrder, err := order.NewOrder(order.SideBuy, "BTC-USD", 50000.0, 2.0)
	require.NoError(t, err)

	// Add buy order
	err = book.AddOrder(buyOrder)
	require.NoError(t, err)

	// Create sell order
	sellOrder, err := order.NewOrder(order.SideSell, "BTC-USD", 50000.0, 1.0)
	require.NoError(t, err)

	// Add sell order
	err = book.AddOrder(sellOrder)
	require.NoError(t, err)

	// Verify order statuses
	assert.Equal(t, order.StatusPartial, buyOrder.Status)
	assert.Equal(t, order.StatusFilled, sellOrder.Status)

	// Verify filled quantities
	assert.Equal(t, 1.0, buyOrder.Filled)
	assert.Equal(t, 1.0, sellOrder.Filled)
}

func TestOrderBook_CancelOrder(t *testing.T) {
	book := orderbook.NewOrderBook("BTC-USD")

	// Create buy order
	buyOrder, err := order.NewOrder(order.SideBuy, "BTC-USD", 50000.0, 1.0)
	require.NoError(t, err)

	// Add buy order
	err = book.AddOrder(buyOrder)
	require.NoError(t, err)

	// Cancel order
	err = book.CancelOrder(buyOrder.ID)
	require.NoError(t, err)

	// Verify order status
	assert.Equal(t, order.StatusCancelled, buyOrder.Status)
}

func TestOrderBook_GetBestPrices(t *testing.T) {
	book := orderbook.NewOrderBook("BTC-USD")

	// Create and add buy order
	buyOrder, err := order.NewOrder(order.SideBuy, "BTC-USD", 50000.0, 1.0)
	require.NoError(t, err)
	err = book.AddOrder(buyOrder)
	require.NoError(t, err)

	// Get best bid
	bestBidPrice, bestBidQty, err := book.GetBestBid()
	require.NoError(t, err)
	assert.Equal(t, 50000.0, bestBidPrice)
	assert.Equal(t, 1.0, bestBidQty)

	// Create and add sell order
	sellOrder, err := order.NewOrder(order.SideSell, "BTC-USD", 50100.0, 1.0)
	require.NoError(t, err)
	err = book.AddOrder(sellOrder)
	require.NoError(t, err)

	// Get best ask
	bestAskPrice, bestAskQty, err := book.GetBestAsk()
	require.NoError(t, err)
	assert.Equal(t, 50100.0, bestAskPrice)
	assert.Equal(t, 1.0, bestAskQty)
}
