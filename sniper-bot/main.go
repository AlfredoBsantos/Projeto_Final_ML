package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
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
	wssURL := os.Getenv("ALCHEMY_WSS_URL")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if wssURL == "" || kafkaBroker == "" || kafkaTopic == "" {
		log.Fatal("ERRO: Variáveis de ambiente faltando no .env")
	}

	kafkaWriter := &kafka.Writer{Addr: kafka.TCP(kafkaBroker), Topic: kafkaTopic, Balancer: &kafka.LeastBytes{}}
	defer kafkaWriter.Close()

	targetContracts := map[common.Address]string{
		common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"): "Uniswap V2",
		common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"): "Uniswap V3",
		common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F"): "Sushiswap",
	}

	for {
		log.Println("Produtor: Tentando conectar ao nó Ethereum...")
		rpcClient, err := rpc.Dial(wssURL)
		if err != nil {
			log.Printf("Falha conexão: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		client := ethclient.NewClient(rpcClient)
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Printf("Falha ao obter ChainID: %v", err)
			rpcClient.Close()
			continue
		}

		headers := make(chan *types.Header)
		sub, err := client.SubscribeNewHead(context.Background(), headers)
		if err != nil {
			log.Printf("Falha subscrição: %v", err)
			rpcClient.Close()
			continue
		}

		log.Println("Produtor: Conexão estabelecida. Coletando dados de blocos.")

	Loop:
		for {
			select {
			case err := <-sub.Err():
				log.Printf("Erro subscrição: %v", err)
				break Loop
			case header := <-headers:
				go func(blockHeader *types.Header) {
					block, err := client.BlockByNumber(context.Background(), blockHeader.Number)
					if err != nil {
						return
					}

					baseFee := "0"
					if block.BaseFee() != nil {
						baseFee = block.BaseFee().String()
					}

					for _, tx := range block.Transactions() {
						if tx.To() == nil {
							continue
						}
						if contractName, ok := targetContracts[*tx.To()]; ok {
							from, _ := types.Sender(types.NewEIP155Signer(chainID), tx)
							data := TransactionData{
								Hash: tx.Hash().Hex(), To: tx.To().Hex(), From: from.Hex(), InputData: "0x" + common.Bytes2Hex(tx.Data()),
								Nonce: tx.Nonce(), GasPrice: tx.GasPrice().String(), GasLimit: tx.Gas(), Value: tx.Value().String(),
								Timestamp: int64(block.Time()), BaseFeePerGas: baseFee,
							}
							jsonData, _ := json.Marshal(data)
							err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{Value: jsonData})
							if err != nil {
								log.Printf("!!! PRODUTOR: Falha ao enviar para Kafka: %v", err)
							} else {
								log.Printf(">>> PRODUTOR: Alvo (%s) no Bloco #%d enviado: %s", contractName, block.NumberU64(), tx.Hash().Hex())
							}
						}
					}
				}(header)
			}
		}
		sub.Unsubscribe()
		rpcClient.Close()
		log.Println("Produtor: Conexão perdida. Reconectando...")
		time.Sleep(5 * time.Second)
	}
}
