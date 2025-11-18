package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	MongoDB  MongoDBConfig
	RabbitMQ RabbitMQConfig
	Worker   WorkerConfig
	Shutdown ShutdownConfig
}

type MongoDBConfig struct {
	URI             string
	Database        string
	Collection      string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

type RabbitMQConfig struct {
	URI           string
	QueueName     string
	MaxRetries    int
	RetryDelay    time.Duration
	PrefetchCount int
}

type WorkerConfig struct {
	ProcessingDelay time.Duration
	ShutdownWait    time.Duration
	Workers         int
}

type ShutdownConfig struct {
	CleanupTimeout time.Duration
}

func Load() *Config {
	return &Config{
		MongoDB: MongoDBConfig{
			URI:             getEnv("MONGO_URI", "mongodb://user:password@mongodb:27017/orders_db?authSource=admin"),
			Database:        getEnv("MONGO_DATABASE", "orders_db"),
			Collection:      getEnv("MONGO_COLLECTION", "orders"),
			MaxPoolSize:     getEnvAsUint64("MONGO_MAX_POOL_SIZE", 100),
			MinPoolSize:     getEnvAsUint64("MONGO_MIN_POOL_SIZE", 10),
			MaxConnIdleTime: getEnvAsDuration("MONGO_MAX_CONN_IDLE_TIME", 30*time.Second),
			ConnectTimeout:  getEnvAsDuration("MONGO_CONNECT_TIMEOUT", 10*time.Second),
		},
		RabbitMQ: RabbitMQConfig{
			URI:           getEnv("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/"),
			QueueName:     getEnv("RABBITMQ_QUEUE_NAME", "orders_queue"),
			MaxRetries:    getEnvAsInt("RABBITMQ_MAX_RETRIES", 5),
			RetryDelay:    getEnvAsDuration("RABBITMQ_RETRY_DELAY", 2*time.Second),
			PrefetchCount: getEnvAsInt("RABBITMQ_PREFETCH_COUNT", 1),
		},
		Worker: WorkerConfig{
			ProcessingDelay: getEnvAsDuration("WORKER_PROCESSING_DELAY", 2*time.Second),
			ShutdownWait:    getEnvAsDuration("WORKER_SHUTDOWN_WAIT", 3*time.Second),
			Workers:         getEnvAsInt("WORKER_POOL_SIZE", 10),
		},
		Shutdown: ShutdownConfig{
			CleanupTimeout: getEnvAsDuration("SHUTDOWN_CLEANUP_TIMEOUT", 5*time.Second),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsUint64(key string, defaultValue uint64) uint64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
