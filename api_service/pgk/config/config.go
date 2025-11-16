package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	MongoDB  MongoDBConfig
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
	Shutdown ShutdownConfig
}

type MongoDBConfig struct {
	URI              string
	Database         string
	Collection       string
	MaxPoolSize      uint64
	MinPoolSize      uint64
	MaxConnIdleTime  time.Duration
	ConnectTimeout   time.Duration
}

type RabbitMQConfig struct {
	URI            string
	QueueName      string
	ExchangeName   string
	MaxRetries     int
	RetryDelay     time.Duration
	PublishTimeout time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type ShutdownConfig struct {
	HTTPTimeout    time.Duration
	CleanupTimeout time.Duration
}

func Load() *Config {
	return &Config{
		MongoDB: MongoDBConfig{
			URI:              getEnv("MONGO_URI", "mongodb://user:password@mongodb:27017/orders_db?authSource=admin"),
			Database:         getEnv("MONGO_DATABASE", "orders_db"),
			Collection:       getEnv("MONGO_COLLECTION", "orders"),
			MaxPoolSize:      getEnvAsUint64("MONGO_MAX_POOL_SIZE", 100),
			MinPoolSize:      getEnvAsUint64("MONGO_MIN_POOL_SIZE", 10),
			MaxConnIdleTime:  getEnvAsDuration("MONGO_MAX_CONN_IDLE_TIME", 30*time.Second),
			ConnectTimeout:   getEnvAsDuration("MONGO_CONNECT_TIMEOUT", 10*time.Second),
		},
		RabbitMQ: RabbitMQConfig{
			URI:            getEnv("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/"),
			QueueName:      getEnv("RABBITMQ_QUEUE_NAME", "orders_queue"),
			ExchangeName:   getEnv("RABBITMQ_EXCHANGE_NAME", ""),
			MaxRetries:     getEnvAsInt("RABBITMQ_MAX_RETRIES", 5),
			RetryDelay:     getEnvAsDuration("RABBITMQ_RETRY_DELAY", 2*time.Second),
			PublishTimeout: getEnvAsDuration("RABBITMQ_PUBLISH_TIMEOUT", 5*time.Second),
		},
		Server: ServerConfig{
			Port:         getEnv("API_PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("API_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("API_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvAsDuration("API_IDLE_TIMEOUT", 60*time.Second),
		},
		Shutdown: ShutdownConfig{
			HTTPTimeout:    getEnvAsDuration("SHUTDOWN_HTTP_TIMEOUT", 10*time.Second),
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
