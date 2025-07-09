// main.go
package main

import (
	"log"
	"os"

	"github.com/DVTcode/podcast_server/config"
	"github.com/DVTcode/podcast_server/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	r := gin.Default()
	routes.SetupRoutes(r, config.DB)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	log.Println("🚀 Server running at http://localhost" + port)
	r.Run(port)
}
