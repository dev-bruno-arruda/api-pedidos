package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/ports"
	"github.com/google/uuid"
)

type OrderService struct {
	repo      ports.OrderRepository
	publisher ports.MessagePublisher
}

func NewOrderService(repo ports.OrderRepository, publisher ports.MessagePublisher) *OrderService {
	return &OrderService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	orderID := uuid.New().String()

	order := &models.Order{
		OrderID:   orderID,
		Product:   req.Product,
		Quantity:  req.Quantity,
		Status:    models.StatusCriado,
		CreatedAt: time.Now(),
	}

	err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar o pedido: %w", err)
	}

	message := models.OrderMessage{
		OrderID: orderID,
		Status:  models.StatusProcessando,
	}
	//tempo para verificar status no banco e andamento do processamento
	time.Sleep(10 * time.Second)

	err = s.publisher.PublishOrderMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	return &models.CreateOrderResponse{
		OrderID: orderID,
		Status:  models.StatusCriado,
	}, nil
}
