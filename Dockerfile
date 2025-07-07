FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Thêm quyền thực thi cho wait-for-it.sh
RUN chmod +x wait-for-it.sh

# Build Go binary
RUN go build -o main ./cmd/main.go

# Dùng wait-for-it.sh trước khi chạy binary
CMD ["./wait-for-it.sh", "db:3306", "--", "./main"]
