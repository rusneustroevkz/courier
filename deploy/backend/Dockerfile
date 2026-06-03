FROM golang:1.25.9-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/backend/main.go

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yml .
COPY .env .

EXPOSE 8080

CMD ["./main"]
