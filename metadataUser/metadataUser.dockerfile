FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o metadata-user ./metadataUser/cmd/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/metadata-user .

EXPOSE 8081
CMD ["./metadata-user"]
