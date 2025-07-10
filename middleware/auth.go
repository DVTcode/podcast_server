package middleware

import (
	"net/http"
	"strings"

	"github.com/DVTcode/podcast_server/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu Authorization header"})
			c.Abort()
			return
		}

		// Tách token khỏi chuỗi "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header không hợp lệ"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc hết hạn"})
			c.Abort()
			return
		}

		// Lưu thông tin vào context để controller dùng
		c.Set("user_id", claims.UserID)
		c.Set("vai_tro", claims.Role)

		c.Next()
	}
}
