# Étape 1 : builder
FROM golang:1.24.2 as builder


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o onlyflick ./cmd/api

# Étape 2 : exécution légère
FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/onlyflick .

ENV PORT=8080
EXPOSE 8080

CMD ["./onlyflick"]
