package ws

import (
	"log"
	"net/http"

	"github.com/DVTcode/podcast_server/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Cân nhắc giới hạn origin khi deploy production
		return true
	},
}

func HandleWebSocket(c *gin.Context) {
	docID := c.Param("id")

	// Xác thực JWT từ query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token"})
		return
	}

	claims, err := utils.VerifyToken(token)
	if err != nil {
		log.Println("Token không hợp lệ:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc hết hạn"})
		return
	}

	userID := claims.UserID
	log.Printf("WebSocket connect: docID=%s by userID=%s\n", docID, userID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade WebSocket:", err)
		return
	}

	// Đăng ký client
	H.Register(docID, conn)

	// Gửi message xác nhận
	if err := conn.WriteMessage(websocket.TextMessage, []byte("WebSocket connected to document: "+docID)); err != nil {
		log.Println("Initial WriteMessage error:", err)
		H.Unregister(docID, conn)
		conn.Close()
		return
	}

	// Lặp để giữ kết nối mở
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket closed: docID=%s, userID=%s, err=%v\n", docID, userID, err)
			break
		}
	}

	// Cleanup khi client ngắt kết nối
	H.Unregister(docID, conn)
	conn.Close()
	log.Printf("WebSocket disconnect: docID=%s, userID=%s\n", docID, userID)
}
