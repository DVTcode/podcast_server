# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code & env
COPY . .

# Build app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

# Copy binary, .env, and wait-for-it
COPY --from=builder /app/main .
COPY --from=builder /app/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

EXPOSE 8080

# Start with wait-for-it
CMD ["/wait-for-it.sh", "mysql.railway.internal:3306", "--", "./main"]
