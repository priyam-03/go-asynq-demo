# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git if needed for dependencies
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app main.go

# Runtime stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

CMD ["./app"]
