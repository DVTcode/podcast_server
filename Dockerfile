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

RUN apk --no-cache add ca-certificates bash coreutils

WORKDIR /root/

# Copy binary, .env, and wait-for-it
COPY --from=builder /app/main ./
COPY --from=builder /app/.env .env
COPY --from=builder /app/wait-for-it.sh /wait-for-it.sh

RUN chmod +x /wait-for-it.sh

# ✅ Tạo lại file credentials từ biến môi trường (được truyền từ Railway)
ENV GOOGLE_CREDENTIALS_B64=""

RUN mkdir -p /root/credentials && \
    if [ -n "$GOOGLE_CREDENTIALS_B64" ]; then \
      echo "$GOOGLE_CREDENTIALS_B64" | base64 -d > /root/credentials/google-credentials.json; \
    else \
      echo "WARNING: GOOGLE_CREDENTIALS_B64 not set"; \
    fi

EXPOSE 8080

# Start app with wait-for-it
CMD ["/wait-for-it.sh", "db:3306", "--", "./main"]
