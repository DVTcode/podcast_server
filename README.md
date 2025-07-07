# 🎧 Podcast Server (Golang + MySQL + Docker)

Đây là backend cho hệ thống quản lý podcast. Dự án sử dụng:
- Go (Gin + GORM)
- MySQL (chạy bằng Docker)
- Docker Compose để chạy môi trường phát triển

---

## 🛠 Hướng dẫn cài đặt

### 1. Clone dự án

```bash
git clone https://github.com/DVTcode/podcast_server.git
cd podcast_server
--- 

### 2. Tạo file cấu hình .env
cp .env.example .env    

// Vì file .env bên dự án tôi là 1 file nhạy cảm không nên upload lên github nhưng chỉ dc cho phép bà lấy env.example rồi tạo .env bên máy bà rồi rồi dán ngược lại vô thôi, nhất định phải có .env trong dự án này

### 3. Khởi động bằng Docker
docker-compose up --build

----
Cấu trúc thư mục
podcast_server/
│
├── cmd/                # Chứa file main.go
├── config/             # Cấu hình DB
├── models/             # Struct GORM
├── routes/             # Định tuyến API
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
