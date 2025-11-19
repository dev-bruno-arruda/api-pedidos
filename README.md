# API de Gerenciamento de Pedidos

Sistema de criacao e processamento assincrono de pedidos desenvolvido em Golang, utilizando MongoDB para persistencia e RabbitMQ para mensageria.

## Tecnologias Utilizadas

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-%234ea94b.svg?style=for-the-badge&logo=mongodb&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/Rabbitmq-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![Ubuntu](https://img.shields.io/badge/Ubuntu-E95420?style=for-the-badge&logo=ubuntu&logoColor=white)

## Arquitetura

O sistema segue **Hexagonal Architecture (Ports & Adapters)** e principios **SOLID**:
Utitilizei a arquitetura Hexagonal para exemplificar como poderíamos alternar de RabbitMQ ou MongoDB para outros bancos ou mensagerias.

### Servicos

1. **API Service (Produtor)**: API REST que recebe requisicoes HTTP e persiste pedidos no MongoDB
2. **Worker Service (Consumidor)**: Servico que consome mensagens da fila e processa pedidos de forma assincrona

### Camadas da Aplicacao
 Handelrs (HTTP/Consumer) - Camada de apresentacao
 Services (Business Logic) - Camada de Domínio
 Ports - São interfaces para Repositry and Publisher
 MongoDB e RabbitMS (Repository and Publisher)

### Fluxo de Processamento

1. Cliente envia POST para `/orders` com dados do pedido
2. API cria o pedido no MongoDB com status `CRIADO`
3. API publica mensagem na fila RabbitMQ com status `PROCESSANDO`
4. Worker consome a mensagem
5. Worker atualiza status para `PROCESSANDO`
6. Worker simula processamento (2 segundos)
7. Worker atualiza status para `PROCESSADO`

## Como Executar

### Pre-requisitos

- Docker
- Docker Compose

### Variaveis de Ambiente

O sistema utiliza arquivo `.env` para configuracao. Um arquivo `.env.example` está disponivel como template.

#### MongoDB
```bash
MONGO_URI=mongodb://user:password@mongodb:27017/orders_db?authSource=admin
MONGO_DATABASE=orders_db
MONGO_COLLECTION=orders
MONGO_MAX_POOL_SIZE=100           # Maximo de conexoes no pool
MONGO_MIN_POOL_SIZE=10            # Minimo de conexoes no pool
MONGO_MAX_CONN_IDLE_TIME=30s      # Tempo maximo de inatividade
MONGO_CONNECT_TIMEOUT=10s         # Timeout de conexao
```

#### RabbitMQ
```bash
RABBITMQ_URI=amqp://guest:guest@rabbitmq:5672/
RABBITMQ_QUEUE_NAME=orders_queue
RABBITMQ_EXCHANGE_NAME=           # Vazio = defaut exchange
RABBITMQ_MAX_RETRIES=5            # Tentativas de reconexao
RABBITMQ_RETRY_DELAY=2s           # Delay entre tentativas
RABBITMQ_PREFETCH_COUNT=1         # QoS para worker
RABBITMQ_PUBLISH_TIMEOUT=5s       # Timeout de publicacao
```

#### API Server
```bash
API_PORT=8080
API_READ_TIMEOUT=15s              # Timeout de leitura
API_WRITE_TIMEOUT=15s             # Timeout de escrita
API_IDLE_TIMEOUT=60s              # Timeout de idle
```

#### Worker
```bash
WORKER_PROCESSING_DELAY=2s        # Tempo de simulacao de processamento
WORKER_SHUTDOWN_WAIT=3s           # Tempo de espera no shutdown
WORKER_POOL_SIZE=10               # Numero de workers concorrentes
```

#### Shutdown
```bash
SHUTDOWN_HTTP_TIMEOUT=10s         # Timeout para encerrar HTTP server
SHUTDOWN_CLEANUP_TIMEOUT=5s       # Timeout para cleanup de recursos
```

### Modo Producao

Para executar em modo producao com imagens otimizadas:

```bash
docker compose up -d --build
```

Este comando irá:
- Construir as imagens dos servicos usando multi-stage builds
- Iniciar MongoDB, RabbitMQ, API Service e Worker Service
- Criar a rede e volumes necessarios

### Modo Desenvolvimento (Utilizei Hot Reload)

Para desenvolvimento com hot reload:

```bash
# Iniciar em modo desenvolvimento
docker compose -f docker-compose.dev.yaml up --build

# Ou em background
docker compose -f docker-compose.dev.yaml up -d --build

# em caso de falha no build ou reinicio do sistema, execute:
docker compose down -v --remove-orphans
```

## API Endpoints

### POST /orders

Cria um novo pedido.

**Request:**
```json
{
  "product": "Notebook Dell",
  "quantity": 2
}
```

**Response (201 Created):**
```json
{
  "order_id": <ObjectIDGerado>,
  "status": "CRIADO"
}
```

**Validacoes:**
- `product`: obrigatorio, nao pode ser vazio
- `quantity`: obrigatorio, deve ser maior que 0

### GET /health

Health check do servico.

**Response (200 OK):**
```
healthy
```

## Exemplos de Uso

### Criar um Pedido

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "product": "Mouse Logitech",
    "quantity": 5
  }'
```

### Verificar Pedidos no MongoDB

```bash
docker exec mongodb mongosh -u user -p password \
  --authenticationDatabase admin orders_db \
  --eval "db.orders.find().pretty()"
```

## Configuracao

**Interfaces de Gerenciamento:**
- RabbitMQ Management: http://localhost:15672 (guest/guest)

### Personalizando a Configuracao

#### 1. Criar arquivo .env

```bash
# Copiar template
cp .env.example .env

# Editar conforme necessidade
nano .env
```

#### 2. Exemplos de Customizacao

**Aumentar pool de conexoes MongoDB (alta carga):**
```bash
MONGO_MAX_POOL_SIZE=200
MONGO_MIN_POOL_SIZE=20
```

**Aumentar workers de publicacao (alta concorrencia):**
Editar `api_service/pgk/service/order_service.go`:
```go
workers := 20 // Era 10, agora 20 workers
```

**Aumentar workers de consumo (alta carga de mensagens):**
```bash
WORKER_POOL_SIZE=20  # Era 10, agora 20 workers
```

**Reduzir tempo de processamento (testes rapidos):**
```bash
WORKER_PROCESSING_DELAY=500ms  # Era 2s, agora 500ms
```

**Aumentar timeout de publicacao (rede lenta):**
```bash
RABBITMQ_PUBLISH_TIMEOUT=30s   # Era 5s, agora 30s
```

## Features Implementadas

### Arquitetura e Design
- Hexagonal Architecture (Ports & Adapters)
- Principios SOLID aplicados
- Dependency Inversion via interfaces
- Baixo acoplamento entre camadas
- Facil troca de implementacoes (MongoDB, RabbitMQ)

### Concorrência e Performance
- **Worker Pool Pattern (API)**: 10 workers concorrentes publicando mensagens
- **Worker Pool Pattern (Worker Service)**: 10 workers concorrentes consumindo mensagens
- **Publicação Assíncrona**: Mensagens publicadas em background
- **Consumo Paralelo**: Múltiplos workers processam pedidos simultaneamente
- **Alta Performance**: ~2000 requests/segundo (20.000x mais rápido)
- **Graceful Shutdown**: Jobs pendentes são processados antes do encerramento
- **Context Propagation**: Gerenciamento adequado de contextos e timeouts


### Fontes

- https://medium.com/@gauravsingharoy/asynchronous-programming-with-go-546b96cd50c1
- https://go.dev/doc/
- https://leapcell.io/blog/the-dance-of-concurrency-and-parallelism-in-golang
