// main.go
package main

import (
	"log"

	"github.com/DVTcode/podcast_server/config"
	"github.com/DVTcode/podcast_server/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.InitDB()
	if err != nil {
		log.Fatal("Không thể kết nối DB: ", err)
	}

	r := gin.Default()
	routes.SetupRoutes(r, db)
	r.Run(":8080")
}
