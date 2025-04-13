FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y gcc libc6-dev

RUN go mod tidy

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o krishaBot .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/krishaBot .

RUN chmod +x krishaBot

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /app/database /app/logs && chmod 777 /app/database /app/logs

CMD ["./krishaBot"]
