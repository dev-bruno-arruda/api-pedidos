package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/models"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/ports"
)

type OrderProcessor struct {
	repo            ports.OrderRepository
	processingDelay time.Duration
}

func NewOrderProcessor(repo ports.OrderRepository, processingDelay time.Duration) *OrderProcessor {
	return &OrderProcessor{
		repo:            repo,
		processingDelay: processingDelay,
	}
}

func (p *OrderProcessor) ProcessOrder(ctx context.Context, message models.OrderMessage) error {
	log.Printf("Iniciando processamento do pedido: %s", message.OrderID)

	order, err := p.repo.FindByOrderID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("pedido não encontrado: %w", err)
	}

	log.Printf("Pedido encontrado: Product=%s, Quantity=%d, Status=%s",
		order.Product, order.Quantity, order.Status)

	err = p.repo.UpdateStatus(ctx, message.OrderID, models.StatusProcessando)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status para PROCESSANDO: %w", err)
	}
	log.Printf("Status atualizado para PROCESSANDO")

	log.Printf("Processando pedido")
	time.Sleep(p.processingDelay)

	err = p.repo.UpdateStatus(ctx, message.OrderID, models.StatusProcessado)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status para PROCESSADO: %w", err)
	}

	log.Printf("Pedido %s processado com sucesso", message.OrderID)

	return nil
}
