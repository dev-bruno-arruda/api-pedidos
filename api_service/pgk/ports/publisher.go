package ports

import (
	"context"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
)

type MessagePublisher interface {
	PublishOrderMessage(ctx context.Context, message models.OrderMessage) error
	Close() error
}
