# --- Fase de Compilación ---
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias
COPY go.mod ./

# Copiar el resto del código
COPY . .

RUN go mod tidy

# Compilar un binario estático optimizado
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-service cmd/api/main.go

# --- Fase de Ejecución ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar el binario desde la fase anterior
COPY --from=builder /api-service .

EXPOSE 8080

CMD ["./api-service"]
