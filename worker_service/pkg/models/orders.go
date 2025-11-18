package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusCriado      = "CRIADO"
	StatusProcessando = "PROCESSANDO"
	StatusProcessado  = "PROCESSADO"
)

type Order struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	Product   string             `bson:"product" json:"product"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type OrderMessage struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
