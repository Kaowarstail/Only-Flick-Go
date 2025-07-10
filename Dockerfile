# Étape 1 : builder
FROM golang:1.24.2 as builder

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application avec optimisations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o onlyflick ./cmd/api

# Étape 2 : exécution légère
FROM debian:bookworm-slim

# Installer les certificats CA et créer un utilisateur non-root
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -r -u 1000 -m -s /bin/bash appuser

WORKDIR /app

# Copier l'exécutable
COPY --from=builder /app/onlyflick .

# Changer le propriétaire
RUN chown appuser:appuser onlyflick

# Utiliser l'utilisateur non-root
USER appuser

# Variables d'environnement
ENV PORT=8080
EXPOSE 8080

# Commande de démarrage
CMD ["./onlyflick"]