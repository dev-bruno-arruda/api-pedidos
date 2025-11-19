package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/ports"
	"github.com/google/uuid"
)

type OrderService struct {
	repo      ports.OrderRepository
	publisher ports.MessagePublisher
	workers   int
	jobQueue  chan asyncJob
	wg        sync.WaitGroup
}

type asyncJob struct {
	ctx     context.Context
	message models.OrderMessage
}

func NewOrderService(repo ports.OrderRepository, publisher ports.MessagePublisher) *OrderService {
	workers := 10
	service := &OrderService{
		repo:      repo,
		publisher: publisher,
		workers:   workers,
		jobQueue:  make(chan asyncJob, 100),
	}

	for i := 0; i < workers; i++ {
		service.wg.Add(1)
		go service.worker(i)
	}

	return service
}

func (s *OrderService) worker(id int) {
	defer s.wg.Done()
	log.Printf("Worker %d iniciado", id)

	for job := range s.jobQueue {
		log.Printf("Worker %d processando job: OrderID=%s", id, job.message.OrderID)

		err := s.publisher.PublishOrderMessage(job.ctx, job.message)
		if err != nil {
			//Obs: aqui poderia ser implementada uma DLQ para não haver descarte de mensagem
			log.Printf("Worker %d: erro ao publicar mensagem para OrderID=%s: %v",
				id, job.message.OrderID, err)
		} else {
			log.Printf("Worker %d: mensagem publicada com sucesso para OrderID=%s",
				id, job.message.OrderID)
		}
	}

	log.Printf("Worker %d finalizado", id)
}

func (s *OrderService) Shutdown() {
	close(s.jobQueue)
	s.wg.Wait()
	log.Println("Todos os workers foram encerrados")
}

func (s *OrderService) CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	orderID := uuid.New().String()

	now := time.Now()
	order := &models.Order{
		OrderID:   orderID,
		Product:   req.Product,
		Quantity:  req.Quantity,
		Status:    models.StatusCriado,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar o pedido: %w", err)
	}

	message := models.OrderMessage{
		OrderID: orderID,
		Status:  models.StatusProcessando,
	}

	workerCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	job := asyncJob{
		ctx:     workerCtx,
		message: message,
	}
	//enfileirando jobs e não bloqear
	select {
	case s.jobQueue <- job:
		log.Printf("Job enfileirado para OrderID=%s", orderID)
	default:
		//Obs: Optei por não descartar nenhuma mensagem usando backpressure, porém, aqui, poderíamos colocar um retorno de 503 para evitar problemas maiores.
		log.Printf("AVISO: Fila de jobs cheia, job para OrderID=%s pode ser processado com atraso", orderID)
		s.jobQueue <- job
	}

	go func() { //Li que dependendo da versão do go, esse cancelamento é redundante(pelo próprio timeout do contexto), porém, deixei aqui para exemplificar
		//eu poderia usar defer para isso também. Mais uma vez, fiz dessa forma apenas para mostrar que também pode ser feito assim.
		<-workerCtx.Done()
		cancel()
	}()

	return &models.CreateOrderResponse{
		OrderID: orderID,
		Status:  models.StatusCriado,
	}, nil
}
