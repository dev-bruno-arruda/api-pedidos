package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(client *mongo.Client, dbName, collectionName string) *OrderRepository {
	return &OrderRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("erro ao criar pedido: %w", err)
	}

	return nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"order_id": orderID}
	update := bson.M{"$set": bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do pedido: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("pedido não encontrado: %s", orderID)
	}

	return nil
}

func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var order models.Order
	filter := bson.M{"order_id": orderID}

	err := r.collection.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pedido não encontrado: %s", orderID)
		}
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	return &order, nil
}

type MongoDBConfig struct {
	URI             string
	Database        string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

func ConnectMongoDB(ctx context.Context, config MongoDBConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize).
		SetMaxConnIdleTime(config.MaxConnIdleTime).
		SetConnectTimeout(config.ConnectTimeout).
		SetServerSelectionTimeout(config.ConnectTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	return client, nil
}
