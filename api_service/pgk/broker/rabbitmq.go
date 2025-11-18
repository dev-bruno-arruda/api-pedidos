package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	exchangeName string
	timeout      time.Duration
}

type PublisherConfig struct {
	URI          string
	QueueName    string
	ExchangeName string
	MaxRetries   int
	RetryDelay   time.Duration
	Timeout      time.Duration
}

func NewRabbitMQPublisher(config PublisherConfig) (*RabbitMQPublisher, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < config.MaxRetries; i++ {
		conn, err = amqp.Dial(config.URI)
		if err == nil {
			break
		}
		log.Printf("Tentativa %d de %d: Falha ao conectar ao RabbitMQ: %v", i+1, config.MaxRetries, err)
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

	log.Println("Conectado ao RabbitMQ com sucesso")

	return &RabbitMQPublisher{
		conn:         conn,
		channel:      channel,
		queueName:    config.QueueName,
		exchangeName: config.ExchangeName,
		timeout:      config.Timeout,
	}, nil
}

func (r *RabbitMQPublisher) PublishOrderMessage(ctx context.Context, message models.OrderMessage) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("contexto cancelado antes de publicar: %w", ctx.Err())
	default:
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	publishCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	err = r.channel.PublishWithContext(
		publishCtx,
		r.exchangeName,
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("contexto cancelado durante publicação: %w", ctx.Err())
		}
		return fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	log.Printf("Mensagem publicada: OrderID=%s, Status=%s", message.OrderID, message.Status)
	return nil
}

func (r *RabbitMQPublisher) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
