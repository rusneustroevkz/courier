FROM golang:1.25.9-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/client/main.go

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yml ./config/
COPY --from=builder /app/deploy/client-backend/.env .

EXPOSE 8080

CMD ["./main"]
