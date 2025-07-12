Sniper Bot Preditivo: Detecção de Anomalias em Transações DeFi com Aprendizagem Não Supervisionada
Autor: Alfredo B. Santos
Curso: Machine Learning
Data: Julho de 2025

1. Visão Geral do Projeto
Este projeto documenta a construção de um ecossistema completo para a análise e detecção de anomalias em transações na blockchain Ethereum. O objetivo final é desenvolver um modelo de Inteligência Artificial capaz de identificar, em tempo real, transações com características de operações de arbitragem ou MEV (Maximal Extractable Value), que se manifestam como anomalias comportamentais em um mercado altamente competitivo.

Diferente de abordagens que buscam decodificar o conteúdo de cada transação, este projeto foca em uma estratégia de Machine Learning Não Supervisionado, ensinando um modelo a aprender o que é um comportamento "normal" e, assim, isolar os "pontos fora da curva" que representam as oportunidades de maior interesse.

A arquitetura final consiste em duas partes principais:

Um Coletor de Dados de alta performance, escrito em Go, que monitora a Ethereum Mainnet 24/7 e salva dados enriquecidos em um arquivo local.

Um Laboratório de Análise e IA, desenvolvido em um Jupyter Notebook com Python, que processa os dados coletados, faz a engenharia de features e treina o modelo de detecção de anomalias.

2. A Arquitetura do Pipeline de Dados
Após uma fase de pesquisa e desenvolvimento que explorou arquiteturas complexas com Kafka e múltiplos microserviços, optou-se por uma abordagem mais robusta e direta para garantir a integridade dos dados e focar no objetivo de Machine Learning.

O pipeline final é um fluxo simplificado e poderoso:

Coletor em Go (Mainnet) -> Arquivo de Log (.jsonl) -> Laboratório de Análise (Jupyter)

Componentes:
Coletor (sniper-bot):

Linguagem: Go, para alta performance e concorrência.

Conexão: Utiliza WebSockets para se conectar a um nó da Ethereum Mainnet via Alchemy.

Monitoramento: Escuta novos blocos (newHeads) para garantir o acesso a dados completos e confirmados.

Filtragem: Foca a coleta em endereços de contratos de alto volume (Roteadores da Uniswap V2 e V3) para garantir a relevância dos dados.

Enriquecimento: Para cada transação de interesse, coleta metadados cruciais como gasPrice, gasLimit, value, e o baseFeePerGas do bloco, além do inputData completo.

Saída: Salva cada transação enriquecida como uma linha em um arquivo mainnet_data.jsonl.

Laboratório de Análise (data-analysis):

Ambiente: Jupyter Notebook rodando em um ambiente Conda (sniper_env).

Ferramentas: Python, com as bibliotecas pandas para manipulação de dados, matplotlib e seaborn para visualização, e scikit-learn para o Machine Learning.

3. A Ciência de Dados: Do Dado à Detecção
O coração do projeto reside no notebook de análise, que executa um pipeline de Machine Learning completo.

3.1. Engenharia de Features Comportamentais
O objetivo não é entender o conteúdo de cada transação, mas sim sua "linguagem corporal". Para isso, criamos as seguintes features a partir dos dados brutos:

priority_fee (Gorjeta de Urgência): Calculada como gasPrice - baseFeePerGas. É a nossa principal pista, indicando o quão desesperado o remetente estava para que sua transação fosse incluída rapidamente. Uma gorjeta alta é um forte indicador de uma operação de MEV/Arbitragem.

input_data_size (Assinatura de Complexidade): O comprimento do inputData. Transações complexas (como swaps em roteadores modernos) possuem um "script" muito maior que transferências simples.

value e gas_limit: O valor em ETH sendo movido e o "orçamento" de gás da transação.

3.2. Modelo de Aprendizagem Não Supervisionada
Como não temos um "gabarito" prévio do que é uma oportunidade, utilizamos uma abordagem não supervisionada.

Algoritmo: Isolation Forest da biblioteca scikit-learn.

Como Funciona: O modelo aprende a estrutura das transações "normais". Ele constrói "árvores de isolamento" e mede quão fácil é isolar um ponto de dados do resto. Pontos que são facilmente isolados são considerados anomalias.

Objetivo: Identificar os outliers no nosso dataset — as transações cujo comportamento (urgência, complexidade, valor) foge drasticamente do padrão.

4. Resultados e Conclusão
Ao aplicar o modelo em um dataset com 8.163 transações reais coletadas da Mainnet, obtivemos os seguintes resultados:

Detecção: O modelo conseguiu classificar com sucesso um subconjunto de transações como anomalias.

Validação: Uma análise estatística dos resultados mostrou que as transações classificadas como anômalas tinham, em média, uma priority_fee e um gas_limit ordens de magnitude maiores que as transações normais.

Conclusão Final: O projeto provou com sucesso que é possível usar técnicas de aprendizagem não supervisionada para detectar anomalias comportamentais no fluxo de transações da Ethereum. Validamos a hipótese de que operações de arbitragem/MEV deixam uma "impressão digital" característica, mesmo quando seu conteúdo é complexo ou ofuscado.

Este trabalho estabelece uma base sólida para a criação de um bot de trading autônomo, onde as previsões deste modelo podem ser usadas como o principal gatilho para a execução de estratégias de arbitragem em tempo real.

5. Como Executar o Projeto
Pré-requisitos:

Go instalado.

Anaconda/Miniconda instalado.

Chave de API da Alchemy.

Passos:

Estrutura: Organize o projeto na estrutura de pastas definida.

Configuração: Crie o arquivo .env na raiz do projeto e adicione sua ALCHEMY_WSS_URL.

Coleta de Dados:

Abra um terminal na pasta sniper-bot.

Rode go mod tidy para instalar as dependências.

Rode go run main.go. Deixe este terminal rodando para coletar os dados.

Análise e Treinamento:

Abra o Anaconda Prompt e ative o ambiente: conda activate sniper_env.

Navegue até a pasta data-analysis.

Inicie o Jupyter: jupyter notebook.

Abra o notebook do projeto e execute as células para carregar os dados do arquivo mainnet_data.jsonl e treinar o modelo.
