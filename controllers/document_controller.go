package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DVTcode/podcast_server/config"
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

	// Bước 1: Khởi tạo tài liệu với trạng thái ban đầu
	doc := models.TaiLieu{
		ID:            id,
		TenFileGoc:    file.Filename,
		DuongDanFile:  publicURL,
		LoaiFile:      ext[1:],
		KichThuocFile: file.Size,
		TrangThai:     "Đã tải lên",
		NguoiTaiLen:   userID,
	}
	if err := db.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không lưu được tài liệu", "details": err.Error()})
		return
	}
	// Bước 3: Trích xuất nội dung
	noiDung, err := services.NormalizeInput(services.InputSource{
		Type:       inputType,
		FileHeader: file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể trích xuất nội dung", "details": err.Error()})
		return
	}

	// Bước 4: Làm sạch nội dung bằng Gemini
	cleanedContent, err := services.CleanTextPipeline(noiDung)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể làm sạch nội dung", "details": err.Error()})
		return
	}

	db.Model(&doc).Updates(map[string]interface{}{
		"TrangThai":        "Đã trích xuất",
		"NoiDungTrichXuat": cleanedContent,
	})

	// Bước 5: Chuyển văn bản thành audio bằng Google TTS
	// Nhận voice và speaking_rate từ form-data
	voice := c.PostForm("voice")
	if voice == "" {
		voice = "vi-VN-Chirp3-HD-Puck"
	}

	rateStr := c.PostForm("speaking_rate")
	rate := 1.0
	if rateStr != "" {
		if parsedRate, err := strconv.ParseFloat(rateStr, 64); err == nil && parsedRate > 0 {
			rate = parsedRate
		}
	}
	audioData, err := services.SynthesizeText(cleanedContent, voice, rate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo audio", "details": err.Error()})
		return
	}
	// Bước 6: Upload audio lên Supabase
	audioURL, err := utils.UploadBytesToSupabase(audioData, id+".mp3", "audio/mp3")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể upload audio", "details": err.Error()})
		return
	}
	now := time.Now()
	// Bước cuối: Hoàn thành
	db.Model(&doc).Updates(map[string]interface{}{
		"TrangThai":    "Hoàn thành",
		"NgayXuLyXong": &now,
	})
	// Tải lại tài liệu để trả về thông tin đầy đủ
	db.Preload("NguoiDung").First(&doc, "id = ?", doc.ID)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Tải lên thành công",
		"tai_lieu":  doc,
		"audio_url": audioURL, // trả về nếu frontend cần phát
	})
}

// GET /api/admin/documents
type TaiLieuStatusDTO struct {
	ID         string `json:"id"`
	TenFileGoc string `json:"ten_file_goc"`
	TrangThai  string `json:"trang_thai"`
	NgayTaiLen string `json:"ngay_tai_len"`
}

func ListDocumentStatus(c *gin.Context) {
	var taiLieus []models.TaiLieu
	var result []TaiLieuStatusDTO
	var total int64

	// Phân trang
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Tìm kiếm theo tên file
	search := c.Query("search")
	query := config.DB.Model(&models.TaiLieu{})

	if search != "" {
		query = query.Where("LOWER(ten_file_goc) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	// Đếm tổng
	query.Count(&total)

	// Lấy dữ liệu
	if err := query.Offset(offset).Limit(limit).Order("ngay_tai_len desc").Find(&taiLieus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách tài liệu", "details": err.Error()})
		return
	}

	// Rút gọn kết quả
	for _, doc := range taiLieus {
		result = append(result, TaiLieuStatusDTO{
			ID:         doc.ID,
			TenFileGoc: doc.TenFileGoc,
			TrangThai:  doc.TrangThai,
			NgayTaiLen: doc.NgayTaiLen.Format("2006-01-02 15:04:05"),
		})
	}

	// Trả về JSON
	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GET /api/admin/documents
type TaiLieuStatusDTO struct {
	ID         string `json:"id"`
	TenFileGoc string `json:"ten_file_goc"`
	TrangThai  string `json:"trang_thai"`
	NgayTaiLen string `json:"ngay_tai_len"`
}

func ListDocumentStatus(c *gin.Context) {
	var taiLieus []models.TaiLieu
	var result []TaiLieuStatusDTO
	var total int64

	// Phân trang
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Tìm kiếm theo tên file
	search := c.Query("search")
	query := config.DB.Model(&models.TaiLieu{})

	if search != "" {
		query = query.Where("LOWER(ten_file_goc) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	// Đếm tổng
	query.Count(&total)

	// Lấy dữ liệu
	if err := query.Offset(offset).Limit(limit).Order("ngay_tai_len desc").Find(&taiLieus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách tài liệu", "details": err.Error()})
		return
	}

	// Rút gọn kết quả
	for _, doc := range taiLieus {
		result = append(result, TaiLieuStatusDTO{
			ID:         doc.ID,
			TenFileGoc: doc.TenFileGoc,
			TrangThai:  doc.TrangThai,
			NgayTaiLen: doc.NgayTaiLen.Format("2006-01-02 15:04:05"),
		})
	}

	// Trả về JSON
	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
