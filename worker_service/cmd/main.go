package main

import (
	"log"
	"os"
	"time"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	rabbitmqURI := os.Getenv("RABBITMQ_URI")

	log.Println("Worker Service iniciado")
	log.Printf("MongoDB URI: %s\n", mongoURI)
	log.Printf("RabbitMQ URI: %s\n", rabbitmqURI)

	// Simula processamento contínuo
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("Worker processando...")

	for range ticker.C {
		log.Println("Worker ainda está vivo e processando...")
	}
}
