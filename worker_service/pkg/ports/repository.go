package ports

import (
	"context"

	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/models"
)

type OrderRepository interface {
	UpdateStatus(ctx context.Context, orderID string, status string) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Order, error)
}
