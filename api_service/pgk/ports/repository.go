package ports

import (
	"context"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID string, status string) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Order, error)
}
