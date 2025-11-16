package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderStatus representa os possiveis estados de um pedido
type OrderStatus string

const (
	StatusCriado      OrderStatus = "CRIADO"
	StatusProcessando OrderStatus = "PROCESSANDO"
	StatusProcessado  OrderStatus = "PROCESSADO"
)

// Order representa um pedido no sistema
type Order struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	Product   string             `bson:"product" json:"product"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Status    OrderStatus        `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// CreateOrderRequest representa a requisicao para criar um pedido
type CreateOrderRequest struct {
	Product  string `json:"product" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

// CreateOrderResponse representa a resposta ao criar um pedido
type CreateOrderResponse struct {
	OrderID string      `json:"order_id"`
	Status  OrderStatus `json:"status"`
}

// OrderMessage representa a mensagem enviada para a fila
type OrderMessage struct {
	OrderID string      `json:"order_id"`
	Status  OrderStatus `json:"status"`
}
