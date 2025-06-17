# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go get ./...

COPY . .
COPY .git .git


RUN apk update
RUN apk add git

RUN git rev-parse --short HEAD

RUN go build -ldflags="-X 'main.Version=$(git rev-parse --short HEAD)'" -o fishstox-web ./cmd/web/main.go

FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /app

COPY --from=builder /app/fishstox-web .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/internal/db/sql/ ./internal/db/sql/

CMD ["./fishstox-web"]
