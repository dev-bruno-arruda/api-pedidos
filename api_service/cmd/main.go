package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	rabbitmqURI := os.Getenv("RABBITMQ_URI")

	log.Println("API Service iniciado")
	log.Printf("MongoDB URI: %s\n", mongoURI)
	log.Printf("RabbitMQ URI: %s\n", rabbitmqURI)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "API Service OK\n")
		fmt.Fprintf(w, "MongoDB: %s\n", mongoURI)
		fmt.Fprintf(w, "RabbitMQ: %s\n", rabbitmqURI)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "healthy")
		log.Printf("Está vivo!")
	})

	log.Println("🌐 Servidor rodando na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
