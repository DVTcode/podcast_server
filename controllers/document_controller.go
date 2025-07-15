package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/DVTcode/podcast_server/models"
	"github.com/DVTcode/podcast_server/services"
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

	// Kiểm tra kích thước (tối đa 20MB)
	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File vượt quá 20MB"})
		return
	}

	// Xác định loại input từ phần mở rộng
	ext := filepath.Ext(file.Filename)
	inputType, err := services.GetInputTypeFromExt(ext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Upload file lên Supabase
	id := uuid.New().String()
	publicURL, err := utils.UploadFileToSupabase(file, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi upload Supabase", "details": err.Error()})
		return
	}

	// Gọi NormalizeInput để trích xuất nội dung
	noiDung, err := services.NormalizeInput(services.InputSource{
		Type:       inputType,
		FileHeader: file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể trích xuất nội dung", "details": err.Error()})
		return
	}

	// Làm sạch nội dung bằng Gemini
	cleanedContent, err := services.CleanTextPipeline(noiDung)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể làm sạch nội dung", "details": err.Error()})
		return
	}

	// Lưu thông tin tài liệu vào DB
	doc := models.TaiLieu{
		ID:               id,
		TenFileGoc:       file.Filename,
		DuongDanFile:     publicURL,
		LoaiFile:         ext[1:], // bỏ dấu chấm
		KichThuocFile:    file.Size,
		TrangThai:        "Đã tải lên",
		NguoiTaiLen:      userID,
		NoiDungTrichXuat: cleanedContent,
	}
	if err := db.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không lưu được tài liệu", "details": err.Error()})
		return
	}

	// Trả về kết quả có preload người dùng
	if err := db.Preload("NguoiDung").First(&doc, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể load thông tin người dùng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Tải lên thành công",
		"tai_lieu":    doc,
		"raw_content": noiDung,
	})
}
