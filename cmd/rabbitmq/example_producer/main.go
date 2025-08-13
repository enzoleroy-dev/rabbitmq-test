package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"rabbitmq-test/config"
	"rabbitmq-test/pkg/rabbitmqx"

	"github.com/shopspring/decimal"
)

// generateLaosDepositMessage creates a sample Laos deposit message
func generateLaosDepositMessage(accountId string) map[string]interface{} {
	amount := decimal.NewFromFloat(float64(rand.Intn(1000000)) / 100)
	return map[string]interface{}{
		"txnId":        "DEP" + time.Now().Format("20060102150405"),
		"paymentRefId": "REF" + time.Now().Format("20060102150405"),
		"status":       "SUCCESS",
		"method":       "DEPOSIT",
		"amount":       amount,
		"currency":     "LAK",
		"timestamp":    time.Now(),
		"payor": map[string]interface{}{
			"number":   "TODO",
			"currency": "TODO",
			"bankCode": "TODO",
		},
		"payee": map[string]interface{}{
			"number":   accountId,
			"currency": "LAK",
			"bankCode": "JDB",
		},
	}
}

// generateLaosWithdrawMessage creates a sample Laos withdrawal message
func generateLaosWithdrawMessage(accountId string) map[string]interface{} {
	amount := decimal.NewFromFloat(float64(rand.Intn(1000000)) / 100)
	return map[string]interface{}{
		"txnId":        "WDR" + time.Now().Format("20060102150405"),
		"paymentRefId": "REF" + time.Now().Format("20060102150405"),
		"status":       "SUCCESS",
		"method":       "WITHDRAW",
		"amount":       amount,
		"currency":     "LAK",
		"timestamp":    time.Now(),
		"payor": map[string]interface{}{
			"number":   accountId,
			"currency": "LAK",
			"bankCode": "JDB",
		},
		"payee": map[string]interface{}{
			"number":   "1023635xxx",
			"currency": "LAK",
			"bankCode": "KBANK",
		},
	}
}

func main() {
	// Load RabbitMQ configuration
	cfg := config.LoadConsumerConfig(config.Env)

	// Create RabbitMQ producer
	producer, err := rabbitmqx.NewProducer(rabbitmqx.Config{
		User:          cfg.RabbitMQUser,
		Password:      cfg.RabbitMQPassword,
		URL:           cfg.RabbitMQURL,
		UseTLS:        cfg.RabbitMQUseTLS,
		TLSCACert:     cfg.RabbitMQTLSCACert,
		TLSCert:       cfg.RabbitMQTLSCert,
		TLSKey:        cfg.RabbitMQTLSKey,
		TLSSkipVerify: cfg.RabbitMQTLSSkipVerify,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Account IDs to test with
	accountIds := []string{"00120010010106019", "70120010010106020"}

	// Get topic patterns from environment variables
	depositTopicPattern := cfg.RabbitLaosDepositTopic
	withdrawTopicPattern := cfg.RabbitLaosWithdrawalTopic

	log.Printf("Using deposit topic pattern: %s", depositTopicPattern)
	log.Printf("Using withdraw topic pattern: %s", withdrawTopicPattern)

	for _, accountId := range accountIds {
		// Replace wildcard with account ID in deposit topic
		depositTopic := strings.Replace(depositTopicPattern, "*", accountId, 1)
		log.Printf("Publishing to deposit topic: %s", depositTopic)

		// Generate and send deposit message
		depositMsg := generateLaosDepositMessage(accountId)
		err = producer.PublishMessage(depositTopic, cfg.RabbitDepositExchangeName, depositMsg, "DEPOSIT")
		if err != nil {
			log.Printf("Failed to publish deposit message for account %s: %v", accountId, err)
		} else {
			log.Printf("Successfully published deposit message for account %s", accountId)
		}

		// Replace wildcard with account ID in withdraw topic
		withdrawTopic := strings.Replace(withdrawTopicPattern, "*", accountId, 1)
		log.Printf("Publishing to withdraw topic: %s", withdrawTopic)

		// Generate and send withdraw message
		withdrawMsg := generateLaosWithdrawMessage(accountId)
		err = producer.PublishMessage(withdrawTopic, cfg.RabbitWithdrawExchangeName, withdrawMsg, "WITHDRAW")
		if err != nil {
			log.Printf("Failed to publish withdraw message for account %s: %v", accountId, err)
		} else {
			log.Printf("Successfully published withdraw message for account %s", accountId)
		}

		// Add a small delay between messages
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("All test messages published successfully")
}
