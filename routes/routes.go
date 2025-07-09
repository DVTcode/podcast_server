package routes

import (
	"github.com/DVTcode/podcast_server/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	api := r.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			controllers.Register(c, db)
		})
	}

}
