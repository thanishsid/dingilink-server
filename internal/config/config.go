package config

import (
	"github.com/caarlos0/env"
)

func Load() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// AllConfig - AllConfig
type Config struct {
	SmtpHost     string `env:"SMTP_HOST"`
	SmtpEmail    string `env:"SMTP_EMAIL"`
	SmtpPassword string `env:"SMTP_PASSWORD"`
	SmtpPort     int    `env:"SMTP_PORT"`

	AwsAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion          string `env:"AWS_REGION"`
	S3Bucket           string `env:"S3_BUCKET"`

	JwtSecretKey string `env:"JWT_SECRET_KEY"`

	DBConnectionString string `env:"DB_CONNECTION_STRING"`

	NatsUrl string `env:"NATS_URL"`

	ServerPort string `env:"SERVER_PORT"`
}
