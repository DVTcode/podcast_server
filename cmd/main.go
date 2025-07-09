package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DVTcode/podcast_server/config"
	"github.com/DVTcode/podcast_server/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Chỉ load .env khi chạy local (Railway không cần)
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		config.LoadEnv()
	}

	// Connect DB
	config.ConnectDB()

	// Setup Gin
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, config.DB)

	// Get port from environment (Railway sets PORT automatically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default cho local
	}

	fmt.Printf("🚀 Server starting on port %s\n", port)

	// Start server
	log.Fatal(r.Run(":" + port))
}
