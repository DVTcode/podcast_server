package routes

import (
	"github.com/DVTcode/podcast_server/controllers"
	"github.com/DVTcode/podcast_server/middleware"
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
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
	user := api.Group("/users")
	{
		user.Use(middleware.AuthMiddleware())
		user.GET("/profile", controllers.GetProfile)
		user.PUT("/profile", controllers.UpdateProfile)
		user.POST("/change-password", controllers.ChangePassword)
	}

	admin := api.Group("/admin")

	{
		admin.Use(middleware.AuthMiddleware(), middleware.DBMiddleware(db)) // ✅ inject db cho nhóm admin
		admin.POST("/documents/upload", controllers.UploadDocument)
		admin.GET("/documents", controllers.ListDocumentStatus)

		admin.POST("/podcasts", controllers.CreatePodcastWithUpload)
		admin.PUT("/podcasts/:id", controllers.UpdatePodcast)

	}

	category := api.Group("/categories")
	{
		category.Use(middleware.AuthMiddleware())
		category.GET("/", controllers.GetDanhMucs)
		category.GET("/:id", controllers.GetDanhMucByID)
		category.POST("/", controllers.CreateDanhMuc)
		category.PUT("/:id", controllers.UpdateDanhMuc)
		category.PUT("/:id/status", controllers.ToggleDanhMucStatus)
	}
	podcast := api.Group("/podcasts")
	{
		podcast.Use(middleware.AuthMiddleware())
		podcast.GET("/", controllers.GetPodcast)
		podcast.GET("/search", controllers.SearchPodcast) // Thêm dòng này
	}
	// Thêm các route khác tại đây
	r.GET("/health", controllers.HealthCheck)
	// Thêm route thực tế tại đây

}
