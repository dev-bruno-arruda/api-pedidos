package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/broker"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/config"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/handler"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/repository"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/service"
)

func main() {
	cfg := config.Load()

	log.Println("API Service Iniciado")
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

	publisher, err := broker.NewRabbitMQPublisher(broker.PublisherConfig{
		URI:          cfg.RabbitMQ.URI,
		QueueName:    cfg.RabbitMQ.QueueName,
		ExchangeName: cfg.RabbitMQ.ExchangeName,
		MaxRetries:   cfg.RabbitMQ.MaxRetries,
		RetryDelay:   cfg.RabbitMQ.RetryDelay,
		Timeout:      cfg.RabbitMQ.PublishTimeout,
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao RabbitMQ: %v", err)
	}
	log.Println("Conectado ao RabbitMQ")

	orderRepo := repository.NewOrderRepository(mongoClient, cfg.MongoDB.Database, cfg.MongoDB.Collection)
	orderService := service.NewOrderService(orderRepo, publisher)
	orderHandler := handler.NewOrderHandler(orderService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "healthy")
		log.Println("Health check: OK")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "API Service OK\n")
		fmt.Fprintf(w, "POST /orders - Criar novo pedido\n")
		fmt.Fprintf(w, "GET /health - Health check\n")
	})

	mux.HandleFunc("/orders", orderHandler.CreateOrder)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor rodando na porta %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	<-quit
	log.Println("Recebido sinal de shutdown")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Shutdown.HTTPTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erro ao fazer shutdown do servidor HTTP: %v", err)
	}
	log.Println("Servidor HTTP encerrado")

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cfg.Shutdown.CleanupTimeout)
	defer cleanupCancel()

	log.Println("Encerrando workers do OrderService...")
	orderService.Shutdown()

	if err := publisher.Close(); err != nil {
		log.Printf("Erro ao fechar conexão com RabbitMQ: %v", err)
	} else {
		log.Println("Conexão RabbitMQ encerrada")
	}

	if err := mongoClient.Disconnect(cleanupCtx); err != nil {
		log.Printf("Erro ao desconectar do MongoDB: %v", err)
	} else {
		log.Println("Conexão MongoDB encerrada")
	}

	log.Println("Shutdown completo")
}
