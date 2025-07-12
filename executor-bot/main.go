package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

type RawTransactionEvent struct {
	Hash string `json:"hash"`
}
type Features struct{ Value, GasLimit, InputDataSize string }
type Prediction struct {
	IsAnomaly int `json:"is_anomaly"`
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Executor: Erro .env: %v", err)
	}

	alchemyHttpURL := os.Getenv("ALCHEMY_HTTPS_URL")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")

	client, err := ethclient.Dial(alchemyHttpURL)
	if err != nil {
		log.Fatalf("Executor: Falha ao conectar ao nó HTTP: %v", err)
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{kafkaBroker}, Topic: kafkaTopic, GroupID: "executor-group"})
	defer kafkaReader.Close()

	log.Println("Executor: Iniciado. Aguardando gatilhos do Coletor...")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigchan:
			log.Println("Executor: Encerrando...")
			return
		default:
			m, err := kafkaReader.ReadMessage(context.Background())
			if err != nil {
				continue
			}

			var event RawTransactionEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				continue
			}

			tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(event.Hash))
			if err != nil || isPending {
				continue
			}

			features := Features{Value: tx.Value().String(), GasLimit: fmt.Sprintf("%d", tx.Gas()), InputDataSize: fmt.Sprintf("%d", len(tx.Data()))}
			jsonData, _ := json.Marshal(features)

			resp, err := http.Post("http://127.0.0.1:5000/predict", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				continue
			}

			var prediction Prediction
			if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			if prediction.IsAnomaly == -1 {
				fmt.Println("*************************************************")
				fmt.Printf("!!! IA SINALIZOU ANOMALIA - POTENCIAL ARBITRAGEM !!!\nHash: %s\n", event.Hash)
				fmt.Println("AÇÃO: Simulação e Execução seriam acionadas aqui.")
				fmt.Println("*************************************************")
				// go executeArbitrage(cfg, tx)
			}
		}
	}
}
