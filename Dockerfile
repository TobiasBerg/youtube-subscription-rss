# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o youtube-subscription-rss .

# ---- Runtime stage ----
FROM alpine:3

WORKDIR /app

COPY --from=builder /app/youtube-subscription-rss .

EXPOSE 8080

ENTRYPOINT ["./youtube-subscription-rss", "server"]
