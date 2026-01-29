package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig      `envPrefix:"APP_"`
	Database DatabaseConfig `envPrefix:"DATABASE_"`
	Redis    RedisConfig    `envPrefix:"REDIS_"`
	RabbitMQ RabbitMQConfig `envPrefix:"RABBITMQ_"`
	Minio    MinioConfig    `envPrefix:"MINIO_"`
	JWT      JWTConfig      `envPrefix:"JWT_"`
	External ExternalConfig `envPrefix:"EXTERNAL_"`
	Logstash LogstashConfig `envPrefix:"LOGSTASH_"`
}

type LogstashConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"50000"`
	Protocol string `env:"PROTOCOL" envDefault:"udp"`
}

type ExternalConfig struct {
	ProductAPIURL string `env:"PRODUCT_API_URL" envDefault:"https://dummyjson.com/products"`
}

type AppConfig struct {
	Name string `env:"NAME" envDefault:"go-boilerplate"`
	Port string `env:"PORT" envDefault:"8080"`
	Mode string `env:"MODE" envDefault:"debug"`
}

type DatabaseConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"5432"`
	User     string `env:"USER" envDefault:"postgres"`
	Password string `env:"PASSWORD"`
	Name     string `env:"NAME"`
	SSLMode  string `env:"SSLMODE" envDefault:"disable"`
}

type RedisConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`
}

type RabbitMQConfig struct {
	URL string `env:"URL" envDefault:"amqp://guest:guest@localhost:5672/"`
}

type MinioConfig struct {
	Endpoint        string `env:"ENDPOINT" envDefault:"localhost:9000"`
	AccessKeyID     string `env:"ACCESS_KEY_ID"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY"`
	UseSSL          bool   `env:"USE_SSL" envDefault:"false"`
	BucketName      string `env:"BUCKET_NAME"`
}

type JWTConfig struct {
	SecretKey string `env:"SECRET_KEY"`
	ExpiresIn int    `env:"EXPIRES_IN" envDefault:"24"` // in hours
}

func LoadConfig() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	return cfg
}
