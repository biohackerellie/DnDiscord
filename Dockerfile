FROM golang:1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -0 ./bin/bot ./cmd/main.go

FROM alpine:latest
COPY --from=builder /app/bin/bot .

ENV APP_ENV=production

CMD ["./bot"]
