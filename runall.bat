@echo off
TITLE Painel de Controle FINAL do Sniper Bot - Ambiente LOCAL

echo [1/4] Forcando a parada e remocao completa do ambiente Docker...
docker-compose down -v

echo [2/4] Iniciando nova infraestrutura Docker...
docker-compose up -d

timeout /t 20 /nobreak > NUL

echo [3/4] Preparando Kafka e Banco de Dados...
docker exec kafka kafka-topics --create --topic mempool-transactions --bootstrap-server kafka:29092 --partitions 1 --replication-factor 1
docker exec -it timescaledb psql -U admin -d mempool_data -c "CREATE TABLE IF NOT EXISTS transactions (hash VARCHAR(66) PRIMARY KEY, to_address VARCHAR(42), from_address VARCHAR(42), nonce BIGINT, gas_price TEXT, gas_limit BIGINT, value TEXT, event_timestamp TIMESTAMPTZ, \"inputData\" TEXT, base_fee_per_gas TEXT);"

echo [4/4] Iniciando bots em novas janelas...
start "Consumidor (Arquivista)" cmd /k "cd data-consumer && go mod tidy && go run main.go"
start "Produtor (Coletor)" cmd /k "cd sniper-bot && go mod tidy && go run main.go"

echo.
echo    ECOSSISTEMA LOCAL INICIADO COM SUCESSO!
pause