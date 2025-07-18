# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash coreutils

WORKDIR /root/

COPY --from=builder /app/main ./
# Copy trực tiếp từ context
COPY wait-for-it.sh /wait-for-it.sh
COPY --from=builder /app/scripts/init-creds.sh ./init-creds.sh
RUN chmod +x wait-for-it.sh init-creds.sh

EXPOSE 8080

# ✅ Gọi init-creds.sh trước khi chạy app
CMD ["bash", "-c", "./init-creds.sh && ./wait-for-it.sh db:3306 -- ./main"]
