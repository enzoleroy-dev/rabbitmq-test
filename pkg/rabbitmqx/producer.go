package rabbitmqx

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

type Config struct {
	User          string
	Password      string
	URL           string
	UseTLS        bool
	TLSCACert     string
	TLSCert       string
	TLSKey        string
	TLSSkipVerify bool
}

// NewProducer creates a new RabbitMQ producer
func NewProducer(cfg Config) (*Producer, error) {
	var conn *amqp.Connection
	var err error

	protocol := "amqp"
	if cfg.UseTLS {
		protocol = "amqps"
	}

	url := fmt.Sprintf("%s://%s:%s@%s", protocol, cfg.User, cfg.Password, cfg.URL)

	if cfg.UseTLS {
		// Configure TLS
		tlsConfig, tlsErr := createTLSConfig(cfg.TLSCACert, cfg.TLSCert, cfg.TLSKey, cfg.TLSSkipVerify)
		if tlsErr != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", tlsErr)
		}

		conn, err = amqp.DialTLS(url, tlsConfig)
	} else {
		conn, err = amqp.Dial(url)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Producer{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

// createTLSConfig creates a TLS configuration for RabbitMQ connection
func createTLSConfig(caCertFile, certFile, keyFile string, skipVerify bool) (*tls.Config, error) {
	// Create a TLS configuration with the certificate of the CA that signed the server's certificate
	rootCAs := x509.NewCertPool()

	if skipVerify {
		return &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: skipVerify,
		}, nil
	}

	if caCertFile != "" {
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		if !rootCAs.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
	}

	// If client certificates are provided, load them
	var certificates []tls.Certificate
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		certificates = append(certificates, cert)
	}

	return &tls.Config{
		RootCAs:      rootCAs,
		Certificates: certificates,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// Close closes the producer connection
func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

// PublishWithdrawMessage publishes a withdrawal message to RabbitMQ using topic exchange
func (p *Producer) PublishWithdrawMessage(topic string, message map[string]interface{}) error {
	// Declare a topic exchange
	exchangeName := "laos_withdraw_exchange"
	err := p.channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare withdraw exchange: %w", err)
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode withdraw message: %w", err)
	}

	// Format JSON for logging
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	log.Printf("Publishing withdraw message to topic '%s':\n%s\n", topic, prettyJSON.String())

	err = p.channel.Publish(
		exchangeName, // exchange
		topic,        // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish withdraw message: %w", err)
	}

	return nil
}

// PublishDepositMessage publishes a deposit message to RabbitMQ using topic exchange
func (p *Producer) PublishDepositMessage(topic string, message map[string]interface{}) error {
	// Declare a topic exchange
	exchangeName := "laos_deposit_exchange"
	err := p.channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare deposit exchange: %w", err)
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode deposit message: %w", err)
	}

	// Format JSON for logging
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	log.Printf("Publishing deposit message to topic '%s':\n%s\n", topic, prettyJSON.String())

	err = p.channel.Publish(
		exchangeName, // exchange
		topic,        // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish deposit message: %w", err)
	}

	return nil
}
