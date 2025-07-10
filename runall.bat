@echo off
TITLE Painel de Controle FINAL do Sniper Bot - Ambiente LOCAL

echo =======================================================
echo.
echo    INICIANDO O ECOSSISTEMA DO SNIPER BOT (vFinal)
echo.
echo =======================================================

REM PASSO 1 (CRUCIAL): Forcar a parada e remocao completa do ambiente anterior
echo [1/5] Forcando a parada e remocao de contêineres e volumes antigos...
docker-compose down -v

echo.
echo [2/5] Iniciando nova infraestrutura Docker...
docker-compose up -d

REM Verifica se o comando anterior foi bem sucedido
if %errorlevel% neq 0 (
    echo.
    echo !!! FALHA CRITICA AO INICIAR O DOCKER. Verifique se o Docker Desktop esta rodando. !!!
    pause
    exit /b
)

echo.
echo [3/5] Aguardando 20 segundos para os servicos estabilizarem...
timeout /t 20 /nobreak > NUL

echo.
echo [4/5] Criando topico no Kafka e tabela no Banco de Dados...
docker exec kafka kafka-topics --create --topic mempool-transactions --bootstrap-server kafka:29092 --partitions 1 --replication-factor 1
docker exec -it timescaledb psql -U admin -d mempool_data -c "CREATE TABLE IF NOT EXISTS transactions (hash VARCHAR(66) PRIMARY KEY, to_address VARCHAR(42), from_address VARCHAR(42), nonce BIGINT, gas_price TEXT, gas_limit BIGINT, value TEXT, event_timestamp TIMESTAMPTZ, \"inputData\" TEXT, base_fee_per_gas TEXT);"

echo.
echo [5/5] Iniciando bots em novas janelas...
start "Consumidor (Arquivista)" cmd /k "cd data-consumer && go run main.go"
start "Produtor (Coletor)" cmd /k "cd sniper-bot && go run main.go"

echo.
echo =======================================================
echo.
echo    ECOSSISTEMA LOCAL INICIADO COM SUCESSO!
echo.
echo =======================================================

pause