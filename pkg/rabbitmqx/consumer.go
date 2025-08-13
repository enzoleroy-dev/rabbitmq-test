package rabbitmqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(cfg Config) (*Consumer, error) {
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

	return &Consumer{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

// Close closes the consumer connection
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// ConsumeDepositQueue starts consuming messages from the deposit queue with wildcard support
func (c *Consumer) ConsumeDepositQueue(topicPattern string, handler func([]byte) error) error {
	// Declare a topic exchange
	exchangeName := "laos_deposit_exchange"
	err := c.channel.ExchangeDeclare(
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

	// Declare a queue with a unique name
	queue, err := c.channel.QueueDeclare(
		"",    // name (empty = auto-generated unique name)
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare deposit queue: %w", err)
	}

	// Bind the queue to the exchange with the topic pattern
	err = c.channel.QueueBind(
		queue.Name,   // queue name
		topicPattern, // routing key (with wildcards)
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind deposit queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register deposit consumer: %w", err)
	}

	log.Printf("Deposit consumer started with topic pattern: %s\n", topicPattern)

	go func() {
		for msg := range msgs {
			// Extract the actual topic from the routing key
			actualTopic := msg.RoutingKey

			// Format JSON for better readability
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, msg.Body, "", "  ")
			messageBody := string(msg.Body)
			if err == nil {
				messageBody = prettyJSON.String()
			}

			// Extract account ID from the topic
			parts := strings.Split(actualTopic, ".")
			accountID := "unknown"
			if len(parts) >= 3 {
				accountID = parts[2]
			}

			log.Printf("\n\n[DEPOSIT] Received message:\nTopic: %s\nAccount ID: %s\nMessage:\n%s\n",
				actualTopic, accountID, messageBody)

			if err := handler(msg.Body); err != nil {
				log.Printf("Error processing deposit message: %v\n", err)
			}
		}
	}()

	return nil
}

// ConsumeWithdrawQueue starts consuming messages from the withdraw queue with wildcard support
func (c *Consumer) ConsumeWithdrawQueue(topicPattern string, handler func([]byte) error) error {
	// Declare a topic exchange
	exchangeName := "laos_withdraw_exchange"
	err := c.channel.ExchangeDeclare(
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

	// Declare a queue with a unique name
	queue, err := c.channel.QueueDeclare(
		"",    // name (empty = auto-generated unique name)
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare withdraw queue: %w", err)
	}

	// Bind the queue to the exchange with the topic pattern
	err = c.channel.QueueBind(
		queue.Name,   // queue name
		topicPattern, // routing key (with wildcards)
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind withdraw queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register withdraw consumer: %w", err)
	}

	log.Printf("Withdraw consumer started with topic pattern: %s\n", topicPattern)

	go func() {
		for msg := range msgs {
			// Extract the actual topic from the routing key
			actualTopic := msg.RoutingKey

			// Format JSON for better readability
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, msg.Body, "", "  ")
			messageBody := string(msg.Body)
			if err == nil {
				messageBody = prettyJSON.String()
			}

			// Extract account ID from the topic
			parts := strings.Split(actualTopic, ".")
			accountID := "unknown"
			if len(parts) >= 3 {
				accountID = parts[2]
			}

			log.Printf("\n\n[WITHDRAW] Received message:\nTopic: %s\nAccount Number: %s\nMessage:\n%s\n",
				actualTopic, accountID, messageBody)

			if err := handler(msg.Body); err != nil {
				log.Printf("Error processing withdraw message: %v\n", err)
			}
		}
	}()

	return nil
}
