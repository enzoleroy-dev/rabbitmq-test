package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type ConsumerConfig struct {
	Log Log

	// RabbitMQ Configuration
	RabbitMQURL      string `env:"RABBITMQ_URL,required"`
	RabbitMQUser     string `env:"RABBITMQ_USER,required"`
	RabbitMQPassword string `env:"RABBITMQ_PASSWORD,required"`

	// RabbitMQ TLS Configuration
	RabbitMQUseTLS   bool   `env:"RABBITMQ_USE_TLS" envDefault:"false"`
	RabbitMQTLSCACert string `env:"RABBITMQ_TLS_CA_CERT" envDefault:""`
	RabbitMQTLSCert   string `env:"RABBITMQ_TLS_CERT" envDefault:""`
	RabbitMQTLSKey    string `env:"RABBITMQ_TLS_KEY" envDefault:""`
	RabbitMQTLSSkipVerify bool `env:"RABBITMQ_TLS_SKIP_VERIFY" envDefault:"false"`

	RabbitLaosDepositTopic    string `env:"RABBITMQ_LAOS_DEPOSIT_TOPIC,required"`
	RabbitLaosWithdrawalTopic string `env:"RABBITMQ_LAOS_WITHDRAWAL_TOPIC,required"`
}

var consumerConfig ConsumerConfig

func LoadConsumerConfig(envPrefix string) ConsumerConfig {
	once.Do(func() {
		opts := env.Options{
			Prefix: prefix(envPrefix),
		}

		var err error
		consumerConfig, err = parseEnv[ConsumerConfig](opts)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
	})

	return consumerConfig
}
