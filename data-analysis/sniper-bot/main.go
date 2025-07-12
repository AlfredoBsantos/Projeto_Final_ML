package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

// Estrutura completa de dados que salvamos no arquivo
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

// Estrutura para enviar as features para a nossa API de IA
type FeaturesForAI struct {
	Value         string `json:"value"`
	GasLimit      uint64 `json:"gas_limit"`
	InputDataSize int    `json:"input_data_size"`
}

// Estrutura para receber a previsão da nossa API de IA
type Prediction struct {
	IsAnomaly int `json:"is_anomaly"`
}

// processTransaction é o coração do nosso bot "agente duplo"
func processTransaction(tx *types.Transaction, block *types.Block, client *ethclient.Client, logFile *os.File) {
	chainID, _ := client.NetworkID(context.Background())
	from, _ := types.Sender(types.NewEIP155Signer(chainID), tx)
	baseFee := "0"
	if block.BaseFee() != nil {
		baseFee = block.BaseFee().String()
	}

	// --- TAREFA 1: AGIR COMO HISTORIADOR ---
	// Prepara os dados completos para salvar no arquivo de log
	dataToSave := TransactionData{
		Hash:          tx.Hash().Hex(),
		To:            tx.To().Hex(),
		From:          from.Hex(),
		InputData:     "0x" + common.Bytes2Hex(tx.Data()),
		Nonce:         tx.Nonce(),
		GasPrice:      tx.GasPrice().String(),
		GasLimit:      tx.Gas(),
		Value:         tx.Value().String(),
		Timestamp:     int64(block.Time()),
		BaseFeePerGas: baseFee,
	}
	jsonDataToSave, _ := json.Marshal(dataToSave)
	if _, err := logFile.WriteString(string(jsonDataToSave) + "\n"); err != nil {
		log.Printf("!!! FALHA AO SALVAR DADO NO ARQUIVO: %v", err)
	} else {
		log.Printf(">>> Dado salvo no arquivo: %s", tx.Hash().Hex())
	}

	// --- TAREFA 2: AGIR COMO PREDADOR ---
	// Prepara as features para consultar a IA em tempo real
	featuresForAI := FeaturesForAI{
		Value:         tx.Value().String(),
		GasLimit:      tx.Gas(),
		InputDataSize: len(tx.Data()),
	}
	jsonDataForAI, _ := json.Marshal(featuresForAI)

	// Consulta a API de IA
	resp, err := http.Post("http://127.0.0.1:5000/predict", "application/json", bytes.NewBuffer(jsonDataForAI))
	if err != nil {
		log.Printf("!!! ERRO ao consultar a API de IA: %v", err)
		return
	}
	defer resp.Body.Close()

	var prediction Prediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return // Ignora erros de decodificação para não poluir o log
	}

	// --- TAREFA 3: AGIR COMO EXECUTOR (COM TRAVA DE SEGURANÇA) ---
	// Se a IA sinalizar uma anomalia, ele reporta a ação
	if prediction.IsAnomaly == -1 { // -1 é o sinal do nosso modelo para "anomalia"
		fmt.Println("*************************************************")
		fmt.Printf("!!! IA SINALIZOU ANOMALIA - POTENCIAL ARBITRAGEM !!!\n")
		fmt.Printf("Hash da Transação Gatilho: %s\n", tx.Hash().Hex())
		fmt.Println("AÇÃO: Simulação e Execução com Flashbots seriam acionadas aqui.")
		fmt.Println("*************************************************")
		// A chamada final para a função 'executeArbitrage' com a lógica da Flashbots viria aqui
		// quando você estiver pronto para operar com capital real.
	}
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Erro ao carregar .env da pasta raiz: %v", err)
	}
	wssURL := os.Getenv("ALCHEMY_WSS_URL")
	if wssURL == "" {
		log.Fatal("ERRO: ALCHEMY_WSS_URL deve estar definido no .env")
	}

	logFile, err := os.OpenFile("mainnet_data.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de dados: %v", err)
	}
	defer logFile.Close()

	targetContracts := map[common.Address]bool{
		common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"): true,
		common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"): true,
	}

	for {
		log.Println("Bot Autônomo: Conectando ao nó Ethereum...")
		rpcClient, err := rpc.Dial(wssURL)
		if err != nil {
			log.Printf("Falha conexão: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		client := ethclient.NewClient(rpcClient)

		headers := make(chan *types.Header)
		sub, err := client.SubscribeNewHead(context.Background(), headers)
		if err != nil {
			log.Printf("Falha subscrição: %v", err)
			rpcClient.Close()
			continue
		}

		log.Println("Conexão estabelecida. Coletando dados e buscando oportunidades...")

	Loop:
		for {
			select {
			case err := <-sub.Err():
				log.Printf("Erro subscrição: %v", err)
				break Loop
			case header := <-headers:
				block, err := client.BlockByNumber(context.Background(), header.Number)
				if err != nil {
					continue
				}
				for _, tx := range block.Transactions() {
					if tx.To() == nil {
						continue
					}
					if _, ok := targetContracts[*tx.To()]; ok {
						go processTransaction(tx, block, client, logFile)
					}
				}
			}
		}
		sub.Unsubscribe()
		rpcClient.Close()
		log.Println("Conexão perdida. Reconectando...")
		time.Sleep(5 * time.Second)
	}
}
