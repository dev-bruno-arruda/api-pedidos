package ports

import (
	"context"

	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/models"
)

type MessageHandler func(ctx context.Context, message models.OrderMessage) error
