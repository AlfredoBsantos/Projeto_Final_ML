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
	_ "github.com/lib/pq" // O driver do PostgreSQL
	"github.com/segmentio/kafka-go"
)

// Estrutura para os dados que recebemos do Kafka.
// Deve ser idêntica à estrutura que o Produtor envia.
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
	// Procura o arquivo .env na pasta pai (a raiz do projeto)
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Erro ao carregar .env da pasta raiz: %v", err)
	}

	// Carrega as configurações do .env
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")

	if dbHost == "" || kafkaBroker == "" || kafkaTopic == "" {
		log.Fatal("ERRO: Variáveis de DB e Kafka devem estar definidas no .env")
	}

	// Monta a string de conexão com o banco de dados
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Consumidor: Falha ao conectar ao DB: %v", err)
	}
	defer db.Close()

	// Configuração do Leitor Kafka
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   kafkaTopic,
		GroupID: "local-storage-group", // Nome do grupo de consumidores
	})
	defer kafkaReader.Close()

	log.Println("Consumidor: Iniciado. Aguardando dados para arquivar...")

	// Canal para lidar com o encerramento gracioso (Ctrl+C)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Loop principal para ler e salvar mensagens
	for {
		select {
		case <-sigchan:
			log.Println("Consumidor: Encerrando...")
			return
		default:
			m, err := kafkaReader.ReadMessage(context.Background())
			if err != nil {
				// Ignora erros de conexão temporários e continua tentando
				continue
			}

			var data TransactionData
			if err := json.Unmarshal(m.Value, &data); err != nil {
				log.Printf("Consumidor: Erro ao decodificar JSON: %v", err)
				continue
			}

			// Comando SQL para inserir os dados na tabela correta
			sqlStatement := `
				INSERT INTO transactions (hash, to_address, from_address, "inputData", event_timestamp, nonce, gas_price, gas_limit, value, base_fee_per_gas) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
				ON CONFLICT (hash) DO NOTHING;`

			_, err = db.Exec(sqlStatement, data.Hash, data.To, data.From, data.InputData, time.Unix(data.Timestamp, 0), data.Nonce, data.GasPrice, data.GasLimit, data.Value, data.BaseFeePerGas)
			if err != nil {
				log.Printf("!!! CONSUMIDOR: Erro ao inserir no DB: %v", err)
			} else {
				log.Printf("[DADO ENRIQUECIDO SALVO] Hash: %s", data.Hash)
			}
		}
	}
}
