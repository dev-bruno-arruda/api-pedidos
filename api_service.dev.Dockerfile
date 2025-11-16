FROM golang:1.24-alpine

# Instala Air para hot reload (versao compativel com Go 1.24)
RUN go install github.com/air-verse/air@v1.61.5

WORKDIR /app

# Copia os arquivos de dependências
COPY api_service/ ./api_service/

# Baixa as dependências
WORKDIR /app/api_service
RUN go mod download

WORKDIR /app

# Define timezone
ENV TZ=America/Sao_Paulo

# Expõe a porta
EXPOSE 8080

# Muda para o diretório do serviço e roda o Air
WORKDIR /app/api_service
CMD ["air"]
