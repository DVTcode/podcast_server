package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/DVTcode/podcast_server/models"
	"github.com/DVTcode/podcast_server/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UploadDocument(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetString("user_id")

	// Nhận file từ form-data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không có file đính kèm"})
		return
	}

	// Kiểm tra kích thước
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File vượt quá 50MB"})
		return
	}

	// Kiểm tra định dạng
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".txt":  true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chỉ cho phép PDF, DOC, DOCX, TXT"})
		return
	}

	// Upload file lên Supabase
	id := uuid.New().String()
	publicURL, err := utils.UploadFileToSupabase(file, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi upload Supabase", "details": err.Error()})
		return
	}

	// Lưu vào DB
	taiLieu := models.TaiLieu{
		ID:            id,
		TenFileGoc:    file.Filename,
		DuongDanFile:  publicURL,
		LoaiFile:      ext[1:], // bỏ dấu chấm
		KichThuocFile: file.Size,
		TrangThai:     "Đã tải lên",
		NguoiTaiLen:   userID,
	}
	if err := db.Create(&taiLieu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không lưu được tài liệu", "details": err.Error()})
		return
	}

	// Truy vấn lại để preload người dùng
	var fullTaiLieu models.TaiLieu
	if err := db.Preload("NguoiDung").First(&fullTaiLieu, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể load thông tin người dùng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Tải lên thành công",
		"tai_lieu": fullTaiLieu,
	})
}
