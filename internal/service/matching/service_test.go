package matching

import (
	"testing"

	"company.com/matchengine/internal/domain/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrder represents test order data
type TestOrder struct {
	side     order.Side
	symbol   string
	price    float64
	quantity float64
}

// createTestOrder is a helper function to create test orders
func createTestOrder(data TestOrder) (*order.Order, error) {
	return order.NewOrder(data.side, data.symbol, data.price, data.quantity)
}

func TestMatchingService(t *testing.T) {
	testCases := []struct {
		name         string
		buyOrder     TestOrder
		sellOrder    TestOrder
		expectedBids int
		expectedAsks int
	}{
		{
			name: "basic order matching",
			buyOrder: TestOrder{
				side:     order.SideBuy,
				symbol:   "BTC-USD",
				price:    50000.0,
				quantity: 1.0,
			},
			sellOrder: TestOrder{
				side:     order.SideSell,
				symbol:   "BTC-USD",
				price:    50000.0,
				quantity: 0.7,
			},
			expectedBids: 1,
			expectedAsks: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service := NewService()

			// Create orders
			buyOrder, err := createTestOrder(tc.buyOrder)
			require.NoError(t, err)

			sellOrder, err := createTestOrder(tc.sellOrder)
			require.NoError(t, err)

			// Execute
			err = service.AddOrder(buyOrder)
			require.NoError(t, err, "failed to add buy order")

			err = service.AddOrder(sellOrder)
			require.NoError(t, err, "failed to add sell order")

			// Verify orderbook state
			book, err := service.GetOrderBook("BTC-USD")
			require.NoError(t, err)

			// Verify order states
			assert.Equal(t, order.StatusPartial, buyOrder.Status)
			assert.Equal(t, order.StatusFilled, sellOrder.Status)
			assert.Equal(t, 0.7, buyOrder.Filled)
			assert.Equal(t, 0.7, sellOrder.Filled)

			// Verify orderbook levels
			assert.Len(t, book.Bids, tc.expectedBids)
			assert.Len(t, book.Asks, tc.expectedAsks)
		})
	}
}

func TestOrderCancellation(t *testing.T) {
	// Setup
	service := NewService()
	orderData := TestOrder{
		side:     order.SideBuy,
		symbol:   "BTC-USD",
		price:    50000.0,
		quantity: 1.0,
	}

	// Create and add order
	createdOrder, err := createTestOrder(orderData)
	require.NoError(t, err)

	err = service.AddOrder(createdOrder)
	require.NoError(t, err)

	// Cancel order
	err = service.CancelOrder(createdOrder.Symbol, createdOrder.ID)
	require.NoError(t, err)

	// Verify cancellation
	assert.Equal(t, createdOrder.Status, order.StatusCancelled)

	// Verify empty orderbook
	book, err := service.GetOrderBook("BTC-USD")
	require.NoError(t, err)
	assert.Empty(t, book.Bids)
}

func TestErrorCases(t *testing.T) {
	testCases := []struct {
		name        string
		testFunc    func(*Service) error
		expectedErr bool
	}{
		{
			name: "invalid symbol orderbook",
			testFunc: func(s *Service) error {
				_, err := s.GetOrderBook("INVALID-PAIR")
				return err
			},
			expectedErr: true,
		},
		{
			name: "cancel order with invalid symbol",
			testFunc: func(s *Service) error {
				return s.CancelOrder("INVALID-PAIR", "some-id")
			},
			expectedErr: true,
		},
		{
			name: "cancel non-existent order",
			testFunc: func(s *Service) error {
				return s.CancelOrder("BTC-USD", "non-existent-id")
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService()
			err := tc.testFunc(service)
			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestMatchingService_ErrorCases(t *testing.T) {
	t.Run("invalid symbol", func(t *testing.T) {
		service := NewService()

		// Try to get non-existent order book
		_, err := service.GetOrderBook("INVALID-PAIR")
		if err == nil {
			t.Error("expected error for invalid symbol, got nil")
		}

		// Try to cancel order for non-existent symbol
		err = service.CancelOrder("INVALID-PAIR", "some-id")
		if err == nil {
			t.Error("expected error for invalid symbol, got nil")
		}
	})

	t.Run("invalid order cancellation", func(t *testing.T) {
		service := NewService()

		// Try to cancel non-existent order
		err := service.CancelOrder("BTC-USD", "non-existent-id")
		if err == nil {
			t.Error("expected error when cancelling non-existent order, got nil")
		}
	})
}

func TestMatchingServiceErrors(t *testing.T) {
	service := NewService()

	// Test canceling non-existent order
	err := service.CancelOrder("BTC-USD", "invalid-id")
	if err == nil {
		t.Error("Expected error when canceling non-existent order")
	}

	// Test getting non-existent order book
	_, err = service.GetOrderBook("invalid-symbol")
	if err == nil {
		t.Error("Expected error when getting non-existent order book")
	}
}

func TestOrderMatching(t *testing.T) {
	service := NewService()

	// Create buy order using helper function
	buyOrder, err := createTestOrder(TestOrder{
		side:     order.SideBuy,
		symbol:   "BTC-USD",
		price:    50000.0,
		quantity: 1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create buy order: %v", err)
	}

	// Adicionar ordem de compra
	err = service.AddOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to add buy order: %v", err)
	}

	// Create sell order using helper function
	sellOrder, err := createTestOrder(TestOrder{
		side:     order.SideSell,
		symbol:   "BTC-USD",
		price:    50000.0,
		quantity: 1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create sell order: %v", err)
	}

	// Adicionar ordem de venda
	err = service.AddOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to add sell order: %v", err)
	}

	// Verificar status das ordens após matching
	if buyOrder.Status != order.StatusPartial {
		t.Errorf("Expected buy order status to be %v, got %v", order.StatusPartial, buyOrder.Status)
	}

	if sellOrder.Status != order.StatusFilled {
		t.Errorf("Expected sell order status to be %v, got %v", order.StatusFilled, sellOrder.Status)
	}

	// Verificar quantidades preenchidas
	if buyOrder.Filled != 1.0 {
		t.Errorf("Expected buy order filled quantity to be 1.0, got %v", buyOrder.Filled)
	}

	if sellOrder.Filled != 1.0 {
		t.Errorf("Expected sell order filled quantity to be 1.0, got %v", sellOrder.Filled)
	}
}
