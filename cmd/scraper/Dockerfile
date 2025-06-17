# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go get ./...

COPY . .
COPY .git .git


RUN go build -o fishstox-scraper ./cmd/scraper/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/fishstox-scraper .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/internal/db/sql/ ./internal/db/sql/

CMD ["./fishstox-scraper"]
