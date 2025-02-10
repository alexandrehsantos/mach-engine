package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"company.com/matchengine/internal/domain/order"
	httphandler "company.com/matchengine/internal/handler/http"
	"company.com/matchengine/internal/service/matching"
)

type MatchingIntegrationSuite struct {
	suite.Suite
	server  *httptest.Server
	service *matching.Service
}

func TestMatchingIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MatchingIntegrationSuite))
}

func (s *MatchingIntegrationSuite) SetupSuite() {
	s.service = matching.NewService()
	handler := httphandler.NewHandler(s.service)
	s.server = httptest.NewServer(handler)
}

func (s *MatchingIntegrationSuite) TearDownSuite() {
	s.server.Close()
}

type OrderRequest struct {
	Symbol   string     `json:"symbol"`
	Side     order.Side `json:"side"`
	Price    float64    `json:"price"`
	Quantity float64    `json:"quantity"`
}

type OrderResponse struct {
	ID        string     `json:"id"`
	Symbol    string     `json:"symbol"`
	Side      order.Side `json:"side"`
	Price     float64    `json:"price"`
	Quantity  float64    `json:"quantity"`
	Status    string     `json:"status"`
	Filled    float64    `json:"filled"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

func (s *MatchingIntegrationSuite) TestOrderMatching() {
	t := s.T()

	// 1. Enviar ordem de compra
	buyOrder := OrderRequest{
		Symbol:   "BTC-USD",
		Side:     order.SideBuy,
		Price:    50000.0,
		Quantity: 1.0,
	}

	buyResp, err := s.sendOrder(buyOrder)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusNew), buyResp.Status)
	assert.Equal(t, 0.0, buyResp.Filled)

	// 2. Verificar orderbook - deve ter 1 ordem de compra
	book, err := s.service.GetOrderBook("BTC-USD")
	require.NoError(t, err)
	assert.Len(t, book.Bids, 1)
	assert.Empty(t, book.Asks)

	// 3. Enviar ordem de venda compatível
	sellOrder := OrderRequest{
		Symbol:   "BTC-USD",
		Side:     order.SideSell,
		Price:    50000.0,
		Quantity: 1.0,
	}

	sellResp, err := s.sendOrder(sellOrder)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusFilled), sellResp.Status)
	assert.Equal(t, 1.0, sellResp.Filled)

	// 4. Verificar orderbook - deve estar vazio após o match
	book, err = s.service.GetOrderBook("BTC-USD")
	require.NoError(t, err)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)

	// 5. Verificar estado final da ordem de compra
	buyOrder2, err := s.getOrder(buyResp.ID)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusFilled), buyOrder2.Status)
	assert.Equal(t, 1.0, buyOrder2.Filled)
}

func (s *MatchingIntegrationSuite) TestOrderCancellation() {
	t := s.T()

	// 1. Enviar ordem de compra
	buyOrder := OrderRequest{
		Symbol:   "BTC-USD",
		Side:     order.SideBuy,
		Price:    50000.0,
		Quantity: 1.0,
	}

	resp, err := s.sendOrder(buyOrder)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusNew), resp.Status)

	// 2. Cancelar a ordem
	err = s.cancelOrder(resp.ID)
	require.NoError(t, err)

	// 3. Verificar que a ordem foi cancelada
	orderResp, err := s.getOrder(resp.ID)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusCancelled), orderResp.Status)

	// 4. Verificar que o orderbook está vazio
	book, err := s.service.GetOrderBook("BTC-USD")
	require.NoError(t, err)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
}

func (s *MatchingIntegrationSuite) sendOrder(order OrderRequest) (*OrderResponse, error) {
	body, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/orders", s.server.URL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

func (s *MatchingIntegrationSuite) getOrder(orderID string) (*OrderResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/orders/%s", s.server.URL, orderID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var order OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *MatchingIntegrationSuite) cancelOrder(orderID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/orders/%s", s.server.URL, orderID), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
