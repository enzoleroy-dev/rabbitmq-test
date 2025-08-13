package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rabbitmq-test/config"
	"rabbitmq-test/pkg/rabbitmqx"
)

type TxnAccount struct {
	Number   string `json:"number"`
	Currency string `json:"currency"`
	BankCode string `json:"bankCode"`
}

type LaosTransactionMessage struct {
	TxnID        string     `json:"txnId"`
	PaymentRefId string     `json:"paymentRefId"`
	Status       string     `json:"status"` // e.g., "SUCCESS", "FAILED"
	Method       string     `json:"method"` // e.g., "DEPOSIT", "WITHDRAW"
	Amount       string     `json:"amount"`
	Currency     string     `json:"currency"`
	Timestamp    time.Time  `json:"timestamp"`
	Payee        TxnAccount `json:"payee"`
	Payor        TxnAccount `json:"payor"`
}

func handleTransactionMessage(msg []byte) error {
	var transactionMsg LaosTransactionMessage
	err := json.Unmarshal(msg, &transactionMsg)
	if err != nil {
		return fmt.Errorf("error parsing transaction message: %v", err)
	}

	// Process the transaction message
	log.Printf("Processing transaction: \n"+
		"  - Transaction ID: %s\n"+
		"  - Payment Ref ID: %s\n"+
		"  - Status: %s\n"+
		"  - Method: %s\n"+
		"  - Amount: %s %s\n"+
		"  - Timestamp: %s\n"+
		"  - Payee Account: %s (Bank: %s)\n"+
		"  - Payor Account: %s (Bank: %s)",
		transactionMsg.TxnID,
		transactionMsg.PaymentRefId,
		transactionMsg.Status,
		transactionMsg.Method,
		transactionMsg.Amount, transactionMsg.Currency,
		transactionMsg.Timestamp.Format(time.RFC3339),
		transactionMsg.Payee.Number, transactionMsg.Payee.BankCode,
		transactionMsg.Payor.Number, transactionMsg.Payor.BankCode)

	return nil
}

func main() {
	// Load RabbitMQ configuration
	cfg := config.LoadConsumerConfig(config.Env)

	// Create RabbitMQ consumer
	consumer, err := rabbitmqx.NewConsumer(rabbitmqx.Config{
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
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Laos deposit consumer
	err = consumer.ConsumeDepositQueue(cfg.RabbitLaosDepositTopic, handleTransactionMessage)
	if err != nil {
		log.Fatalf("Failed to set up deposit consumer: %v", err)
	}

	// Laos withdrawal consumer
	err = consumer.ConsumeWithdrawQueue(cfg.RabbitLaosWithdrawalTopic, handleTransactionMessage)
	if err != nil {
		log.Fatalf("Failed to set up withdraw consumer: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("RabbitMQ consumer started. Press CTRL+C to exit")
	<-sigCh
	log.Println("Shutting down RabbitMQ consumer...")
}
