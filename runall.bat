@echo off
TITLE Painel de Controle - Bot de IA (Pipeline Simplificado)

echo =======================================================
echo.
echo    INICIANDO O ECOSSISTEMA DE IA SIMPLIFICADO
echo.
echo =======================================================

echo [PASSO 1 de 2] Iniciando o Servidor de IA (O Cerebro)...
REM Abre uma nova janela, ativa o conda, entra na pasta de análise e roda a API
start "Cerebro da IA (Python API)" cmd /k "conda activate sniper_env && cd data-analysis && python api_server.py"

echo.
echo    Aguardando 10 segundos para a IA carregar o modelo...
timeout /t 10 /nobreak > NUL

echo.
echo [PASSO 2 de 2] Iniciando o Bot Coletor/Executor (Go)...
REM Abre uma nova janela, entra na pasta do bot e o inicia
start "Bot Coletor/Executor (Go)" cmd /k "cd sniper-bot && go run main.go"

echo.
echo =======================================================
echo.
echo    SISTEMA COMPLETO INICIADO!
echo.
echo    Monitore as 2 novas janelas de terminal.
echo.
echo =======================================================

pause