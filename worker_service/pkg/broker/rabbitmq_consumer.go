package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/logger"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/models"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
	workers   int
}

type ConsumerConfig struct {
	URI           string
	QueueName     string
	MaxRetries    int
	RetryDelay    time.Duration
	PrefetchCount int
	Workers       int
}

func NewRabbitMQConsumer(config ConsumerConfig) (*RabbitMQConsumer, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < config.MaxRetries; i++ {
		conn, err = amqp.Dial(config.URI)
		if err == nil {
			break
		}
		logger.Warnf("Tentativa %d de %d: Falha ao conectar ao RabbitMQ: %v", i+1, config.MaxRetries, err)
		time.Sleep(config.RetryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ após %d tentativas: %w", config.MaxRetries, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal: %w", err)
	}

	_, err = channel.QueueDeclare(
		config.QueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao declarar fila: %w", err)
	}

	err = channel.Qos(
		config.PrefetchCount,
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	logger.Info("Conectado ao RabbitMQ com sucesso")

	return &RabbitMQConsumer{
		conn:      conn,
		channel:   channel,
		queueName: config.QueueName,
		workers:   config.Workers,
	}, nil
}

func (c *RabbitMQConsumer) StartConsuming(ctx context.Context, handler ports.MessageHandler) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("falha ao registrar consumer: %w", err)
	}

	logger.Infof("Iniciando %d workers para processar mensagens...", c.workers)

	var wg sync.WaitGroup
	notifyClose := make(chan *amqp.Error)
	c.channel.NotifyClose(notifyClose)

	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.worker(ctx, i, msgs, handler, &wg)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelado, aguardando workers finalizarem...")
		wg.Wait()
		logger.Info("Todos os workers finalizados")
		return ctx.Err()

	case err := <-notifyClose:
		logger.Errorf("Conexão com RabbitMQ fechada: %v", err)
		wg.Wait()
		return fmt.Errorf("conexão com RabbitMQ perdida: %w", err)
	}
}

func (c *RabbitMQConsumer) worker(ctx context.Context, id int, msgs <-chan amqp.Delivery, handler ports.MessageHandler, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.WorkerInfo(id, "iniciado")

	for {
		select {
		case <-ctx.Done():
			logger.WorkerInfo(id, "encerrando...")
			return

		case msg, ok := <-msgs:
			if !ok {
				logger.WorkerInfo(id, "canal de mensagens fechado")
				return
			}

			logger.WorkerInfof(id, "processando mensagem: OrderID=%s", extractOrderID(msg.Body))

			if err := c.processMessage(ctx, msg, handler); err != nil {
				logger.WorkerErrorf(id, "erro ao processar mensagem: %v", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
				logger.WorkerInfo(id, "mensagem processada com sucesso")
			}
		}
	}
}

func extractOrderID(body []byte) string {
	var orderMsg models.OrderMessage
	if err := json.Unmarshal(body, &orderMsg); err != nil {
		return "unknown"
	}
	return orderMsg.OrderID
}

func (c *RabbitMQConsumer) processMessage(ctx context.Context, msg amqp.Delivery, handler ports.MessageHandler) error {
	var orderMsg models.OrderMessage

	err := json.Unmarshal(msg.Body, &orderMsg)
	if err != nil {
		return fmt.Errorf("erro ao deserializar mensagem: %w", err)
	}

	logger.Infof("Mensagem recebida: OrderID=%s, Status=%s", orderMsg.OrderID, orderMsg.Status)

	err = handler(ctx, orderMsg)
	if err != nil {
		return fmt.Errorf("erro no handler: %w", err)
	}

	return nil
}

func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
