package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/DVTcode/podcast_server/config"
	"github.com/DVTcode/podcast_server/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // ✅ Thêm dòng này để dùng godotenv
)

func main() {
	// ✅ Chỉ load .env khi không chạy Docker (tức là chạy local)
	if os.Getenv("DOCKER_ENV") != "true" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("❌ Load .env failed: %v", err)
		}
	}

	// Connect DB
	config.ConnectDB()

	// Setup Gin
	r := gin.Default()

	// ✅ Bổ sung cấu hình CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "https://podcastserver-production.up.railway.app"}, // 👈 Thay bằng domain frontend thật
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes
	routes.SetupRoutes(r, config.DB)

	// Get port from environment (Railway sets PORT automatically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default local
	}

	fmt.Printf("🚀 Server starting on port %s\n", port)

	// Start server
	log.Fatal(r.Run(":" + port))
}
