package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/broker"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/config"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/repository"
	"github.com/dev-bruno-arruda/api-pedidos/worker_service/pkg/service"
)

func main() {
	cfg := config.Load()

	log.Println("Worker Service Iniciado")
	log.Printf("MongoDB URI: %s", cfg.MongoDB.URI)
	log.Printf("RabbitMQ URI: %s", cfg.RabbitMQ.URI)

	mongoClient, err := repository.ConnectMongoDB(context.Background(), repository.MongoDBConfig{
		URI:             cfg.MongoDB.URI,
		Database:        cfg.MongoDB.Database,
		MaxPoolSize:     cfg.MongoDB.MaxPoolSize,
		MinPoolSize:     cfg.MongoDB.MinPoolSize,
		MaxConnIdleTime: cfg.MongoDB.MaxConnIdleTime,
		ConnectTimeout:  cfg.MongoDB.ConnectTimeout,
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	log.Println("Conectado ao MongoDB")

	consumer, err := broker.NewRabbitMQConsumer(broker.ConsumerConfig{
		URI:           cfg.RabbitMQ.URI,
		QueueName:     cfg.RabbitMQ.QueueName,
		MaxRetries:    cfg.RabbitMQ.MaxRetries,
		RetryDelay:    cfg.RabbitMQ.RetryDelay,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
		Workers:       cfg.Worker.Workers,
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao RabbitMQ: %v", err)
	}
	log.Println("Conectado ao RabbitMQ")

	orderRepo := repository.NewOrderRepository(mongoClient, cfg.MongoDB.Database, cfg.MongoDB.Collection)
	orderProcessor := service.NewOrderProcessor(orderRepo, cfg.Worker.ProcessingDelay)

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		err := consumer.StartConsuming(ctx, orderProcessor.ProcessOrder)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case <-quit:
		log.Println("Recebido sinal de shutdown")
	case err := <-errChan:
		log.Printf("Erro no consumer: %v", err)
	}

	cancel()
	log.Println("Consumer parado")

	log.Println("Aguardando processamento de mensagens pendentes")
	time.Sleep(cfg.Worker.ShutdownWait)

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cfg.Shutdown.CleanupTimeout)
	defer cleanupCancel()

	if err := consumer.Close(); err != nil {
		log.Printf("Erro ao fechar conexão com RabbitMQ: %v", err)
	} else {
		log.Println("Conexão RabbitMQ encerrada")
	}

	if err := mongoClient.Disconnect(cleanupCtx); err != nil {
		log.Printf("Erro ao desconectar do MongoDB: %v", err)
	} else {
		log.Println("Conexão MongoDB encerrada")
	}

	log.Println("Worker encerrado")
}
