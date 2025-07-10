package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type TransactionData struct {
	Hash          string `json:"hash"`
	To            string `json:"to"`
	From          string `json:"from"`
	Nonce         uint64 `json:"nonce"`
	GasPrice      string `json:"gasPrice"`
	GasLimit      uint64 `json:"gasLimit"`
	Value         string `json:"value"`
	Timestamp     int64  `json:"timestamp"`
	InputData     string `json:"inputData"`
	BaseFeePerGas string `json:"baseFeePerGas"`
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Erro ao carregar .env da pasta raiz: %v", err)
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	kafkaBroker := os.Getenv("KAFKA_BROKER")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Consumidor: Falha ao conectar ao DB: %v", err)
	}
	defer db.Close()

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{kafkaBroker}, Topic: "mempool-transactions", GroupID: "local-storage-group"})
	defer kafkaReader.Close()

	log.Println("Consumidor: Iniciado. Aguardando dados...")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigchan:
			log.Println("Consumidor: Encerrando...")
			return
		default:
			m, err := kafkaReader.ReadMessage(context.Background())
			if err != nil {
				continue
			}
			var data TransactionData
			if err := json.Unmarshal(m.Value, &data); err != nil {
				continue
			}

			sqlStatement := `INSERT INTO transactions (hash, to_address, from_address, "inputData", event_timestamp, nonce, gas_price, gas_limit, value, base_fee_per_gas) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (hash) DO NOTHING;`

			_, err = db.Exec(sqlStatement, data.Hash, data.To, data.From, data.InputData, time.Unix(data.Timestamp, 0), data.Nonce, data.GasPrice, data.GasLimit, data.Value, data.BaseFeePerGas)
			if err != nil {
				log.Printf("!!! CONSUMIDOR: Erro ao inserir no DB: %v", err)
			} else {
				log.Printf("[DADO ENRIQUECIDO SALVO] Hash: %s", data.Hash)
			}
		}
	}
}
