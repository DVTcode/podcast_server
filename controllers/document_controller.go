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

	// Đường dẫn tạm local để đọc file sau khi upload (nếu cần)
	// Do file không còn local nên bạn cần mở lại file tạm từ fileHeader nếu cần xử lý trực tiếp
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file đã upload"})
		return
	}
	defer f.Close()

	// Trích xuất nội dung văn bản từ mọi loại file
	noiDung, err := services.NormalizeInput(services.InputSource{
		Type:       services.InputDOCX,
		FileHeader: file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể trích xuất nội dung", "details": err.Error()})
		return
	}

	// Lưu vào DB
	doc := models.TaiLieu{
		ID:               id,
		TenFileGoc:       file.Filename,
		DuongDanFile:     publicURL,
		LoaiFile:         ext[1:],
		KichThuocFile:    file.Size,
		TrangThai:        "Đã tải lên",
		NguoiTaiLen:      userID,
		NoiDungTrichXuat: noiDung,
	}
	if err := db.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không lưu được tài liệu", "details": err.Error()})
		return
	}

	// Preload người dùng để trả kết quả chi tiết
	if err := db.Preload("NguoiDung").First(&doc, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể load thông tin người dùng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Tải lên thành công",
		"tai_lieu": doc,
	})
}
