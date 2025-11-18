package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

const (
    StatusCriado       = "CRIADO"
    StatusProcessando  = "PROCESSANDO"
    StatusProcessado   = "PROCESSADO"
)

type Order struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    OrderID   string             `json:"order_id" bson:"order_id"`
    Product   string             `json:"product" bson:"product"`
    Quantity  int                `json:"quantity" bson:"quantity"`
    Status    string             `json:"status" bson:"status"`
    CreatedAt time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type OrderMessage struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
}

type CreateOrderRequest struct {
    Product  string `json:"product"`
    Quantity int    `json:"quantity"`
}

type CreateOrderResponse struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
}