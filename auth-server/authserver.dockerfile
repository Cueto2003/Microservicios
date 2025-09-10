# Etapa 1: build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar dependencias y descargar módulos
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del proyecto
COPY . .

# Compilar el binario y asegurarlo en /bin/auth-server
RUN go build -o /bin/auth-server ./auth-server/cmd/main.go

# Etapa 2: imagen final
FROM alpine:3.18

WORKDIR /app

# Copiar solo el binario, no directorios
COPY --from=builder /bin/auth-server /app/auth-server

EXPOSE 8082

CMD ["/app/auth-server"]
