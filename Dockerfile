FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make bash && \
    go install github.com/swaggo/swag/cmd/swag@latest
    
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
    
RUN swag init -g ./cmd/subscriptions-api/main.go -o ./docs

# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/subscriptions-api
RUN GOOS=linux go build -o /app/main ./cmd/subscriptions-api

FROM alpine:3.22
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs /app/docs

EXPOSE 8080

ENTRYPOINT ["/app/main"]
