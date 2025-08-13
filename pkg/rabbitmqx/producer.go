package rabbitmqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

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
	TLSEnabled    bool
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
	if cfg.TLSEnabled {
		protocol = "amqps"
	}

	url := fmt.Sprintf("%s://%s:%s@%s", protocol, cfg.User, cfg.Password, cfg.URL)

	if cfg.TLSEnabled {
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

// Close closes the producer connection
func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

// PublishMessage publishes a message to RabbitMQ using topic exchange
// This is a generic function that can be used for any type of message
func (p *Producer) PublishMessage(exchangeName string, topic string, message map[string]interface{}, messageType string) error {
	// Declare a topic exchange
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
		return fmt.Errorf("failed to declare %s exchange: %w", messageType, err)
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode %s message: %w", messageType, err)
	}

	// Format JSON for logging
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	log.Printf("Publishing %s message to topic '%s':\n%s\n", messageType, topic, prettyJSON.String())

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
		return fmt.Errorf("failed to publish %s message: %w", messageType, err)
	}

	return nil
}
