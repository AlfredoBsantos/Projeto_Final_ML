# Sniper Bot Preditivo: Detecção de Anomalias em Transações DeFi

**Autor:** Alfredo B. Santos  
**Status:** Projeto Concluído (Fase de Pesquisa e Desenvolvimento)

## 1. Visão Geral do Projeto

Este projeto documenta a concepção, arquitetura, desenvolvimento e implementação de um ecossistema completo para a **detecção de anomalias em transações na blockchain Ethereum**. A hipótese central é que operações de arbitragem e MEV (Maximal Extractable Value), por sua natureza competitiva e urgente, exibem características comportamentais distintas que podem ser identificadas por modelos de Machine Learning não supervisionado.

O sistema foi construído para ser uma plataforma de pesquisa quantitativa, com foco em um pipeline de dados robusto e um laboratório de análise flexível para testar e validar estratégias de trading algorítmico.

### Tecnologias Utilizadas
* **Coleta de Dados:** Go (Golang)
* **Análise e IA:** Python, Pandas, Scikit-learn, Jupyter Notebook
* **Infraestrutura (Prototipagem):** Docker, Kafka, PostgreSQL
* **Conexão Blockchain:** Alchemy (Nó Ethereum Mainnet)

---

## 2. Arquitetura Final do Projeto (MVP)

Após uma fase de desenvolvimento que explorou uma arquitetura de microsserviços com Kafka, a estratégia foi pivotada para um modelo **simplificado e mais robusto** para o ambiente de pesquisa, eliminando pontos de falha de infraestrutura e focando na qualidade dos dados e na prototipagem da IA.

O pipeline final é um fluxo direto e eficiente:

```
Go Collector (Mainnet) -> Arquivo de Log (.jsonl) -> Laboratório de Análise (Python/Jupyter)
```

### Componentes

#### 🔹 Coletor de Dados (`sniper-bot`)
Um serviço de alta performance escrito em **Go**, responsável por:
- **Conectar-se** a um nó da Ethereum Mainnet via WebSockets.
- **Monitorar** novos blocos (`newHeads`) em tempo real.
- **Filtrar** transações destinadas a uma lista pré-definida de contratos de alto volume (ex: Roteadores da Uniswap).
- **Enriquecer** os dados, capturando metadados cruciais como `gasPrice`, `baseFeePerGas`, `value` e o `inputData` completo.
- **Salvar** cada transação de interesse como uma nova linha em um arquivo local `mainnet_data.jsonl`, garantindo uma coleta de dados persistente e resiliente.

#### 🔹 Laboratório de Análise e IA (`data-analysis`)
Um ambiente de ciência de dados autocontido, utilizando um **Jupyter Notebook**, que realiza o fluxo completo de Machine Learning:
- **Carregamento:** Lê os dados diretamente do arquivo `.jsonl` gerado pelo coletor.
- **Engenharia de Features:** Processa os dados brutos para criar "pistas" comportamentais.
- **Modelagem:** Aplica algoritmos de aprendizagem não supervisionada para encontrar padrões.
- **Análise e Visualização:** Gera relatórios estatísticos e gráficos para interpretar os resultados do modelo.

---

## 3. A Ciência de Dados: Da Análise à Detecção de Anomalias

O coração do projeto é o processo de transformar dados em inteligência.

### 3.1. Engenharia de Features
As seguintes features foram criadas para descrever o "comportamento" de cada transação, em vez de seu conteúdo:

* `priority_fee`: Calculada como `gasPrice - baseFeePerGas`. É nosso principal indicador de **urgência**.
* `input_data_size`: O tamanho do `inputData`. Um forte indicador de **complexidade** da transação.
* `value` e `gasLimit`: O valor em ETH da transação e seu "orçamento" de gás.

### 3.2. Modelo Não Supervisionado
Dada a natureza do problema (encontrar eventos raros e "estranhos" sem um gabarito), a abordagem escolhida foi a de **Detecção de Anomalias**.

* **Algoritmo:** **Isolation Forest** (`sklearn.ensemble.IsolationForest`).
* **Lógica:** O modelo aprende o que é o comportamento "normal" da grande maioria das transações e, em seguida, mede quão fácil é "isolar" cada ponto de dados. Transações que são facilmente isoladas são classificadas como **anomalias**.

---

## 4. Resultados e Conclusão

A aplicação do modelo em um dataset com milhares de transações reais coletadas da Mainnet validou com sucesso a hipótese inicial:

> O modelo não supervisionado foi capaz de identificar um cluster de transações anômalas. Uma análise estatística subsequente provou que essas anomalias possuíam, em média, uma **`priority_fee` e um `gas_limit` significativamente maiores** que as transações normais, confirmando que o modelo está, de fato, detectando as operações de maior urgência e complexidade, que são os alvos mais prováveis para uma estratégia de arbitragem.

Este projeto conclui com sucesso a criação de uma plataforma de análise de ponta a ponta, desde a coleta de dados brutos da blockchain até o treinamento de um modelo de IA capaz de extrair insights valiosos do mercado.

---

## 5. Como Executar o Projeto (Ambiente Local)

**Pré-requisitos:**
- Go (v1.20+)
- Anaconda com um ambiente Python 3.10+
- Chave de API da Alchemy

**Passos:**
1.  **Configuração:** Crie um arquivo `.env` na raiz do projeto com sua `ALCHEMY_WSS_URL`.
2.  **Coleta de Dados:**
    ```bash
    # Abra um terminal
    cd sniper-bot/
    go mod tidy
    go run main.go
    ```
    *Deixe este terminal rodando para coletar dados. Um arquivo `mainnet_data.jsonl` será criado.*
3.  **Análise e Treinamento da IA:**
    * Inicie o Jupyter Notebook.
    * Abra o notebook na pasta `data-analysis`.
    * Execute as células para carregar os dados do `mainnet_data.jsonl` e realizar a análise.
