package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Driver do PostgreSQL
)

func main() {
	// Vamos colocar a string de conexão diretamente aqui para eliminar o .env como uma variável
	// Usando as credenciais do nosso ambiente Docker local
	connStr := "postgres://admin:admin@localhost:5432/mempool_data?sslmode=disable"

	log.Println("Tentando conectar ao banco de dados...")

	// Abre a conexão
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("!!! ERRO AO ABRIR A CONEXÃO: %v", err)
	}
	defer db.Close()

	// Tenta fazer um "ping" para verificar se a conexão está realmente viva
	err = db.Ping()
	if err != nil {
		log.Fatalf("!!! ERRO NO PING DO BANCO DE DADOS: %v", err)
	}

	// Se chegou até aqui, a conexão foi um sucesso absoluto
	fmt.Println("========================================")
	log.Println(">>> SUCESSO! Conexão com o banco de dados estabelecida com sucesso!")
	fmt.Println("========================================")
}
