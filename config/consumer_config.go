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
	RabbitMQTLSEnabled    bool   `env:"RABBITMQ_TLS_ENABLED" envDefault:"false"`
	RabbitMQTLSSkipVerify bool   `env:"RABBITMQ_TLS_SKIP_VERIFY" envDefault:"false"`
	RabbitMQTLSCACert     string `env:"RABBITMQ_TLS_CA_CERT" envDefault:""`
	RabbitMQTLSCert       string `env:"RABBITMQ_TLS_CERT" envDefault:""`
	RabbitMQTLSKey        string `env:"RABBITMQ_TLS_KEY" envDefault:""`

	// RabbitMQ Exchange Names
	RabbitDepositExchangeName  string `env:"RABBITMQ_DEPOSIT_EXCHANGE_NAME,required"`
	RabbitWithdrawExchangeName string `env:"RABBITMQ_WITHDRAW_EXCHANGE_NAME,required"`

	// RabbitMQ Topic Names
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
