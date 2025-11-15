# ====================
# BUILD STAGE
# ====================
FROM golang:1.24-alpine AS builder

ARG APP_PATH=worker_service

WORKDIR /build

# Copia toda a pasta da aplicação
COPY ${APP_PATH} ./

# Baixa as dependências
RUN go mod download

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app_bin ./cmd/main.go


# ====================
# FINAL STAGE
# ====================
FROM alpine:latest

# Define o fuso horário (boa prática)
ENV TZ=America/Sao_Paulo
RUN apk add --no-cache tzdata

# Cria um diretório de trabalho limpo
WORKDIR /root/

# Copia o binário compilado do estágio anterior para o contêiner final
COPY --from=builder /app_bin .

# Comando para rodar a aplicação worker
CMD ["./app_bin"]