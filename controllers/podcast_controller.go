package controllers

import (
	"fmt"
	"net/http"
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

func GetPodcast(c *gin.Context) {
	var podcasts []models.Podcast
	var total int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	search := c.Query("search")
	status := c.Query("status")
	categoryID := c.Query("category")
	sort := c.DefaultQuery("sort", "date")
	query := config.DB.Model(&models.Podcast{})
	//lọc danh podcast theo danh mục, sắp xếp từ A đến Z
	role, _ := c.Get("vai_tro")
	if role != "admin" {
		query = query.Where("trang_thai = ?", "Bật") // Đổi từ "kich_hoat" sang "trang_thai"
	}
	if search != "" {
		query = query.Where("LOWER(tieu_de) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" && role == "admin" {
		switch status {
		case "Bật":
			query = query.Where("trang_thai = ?", "Bật") // Sử dụng đúng trường "trang_thai"
		case "Tắt":
			query = query.Where("trang_thai = ?", "Tắt") // Sử dụng đúng trường "trang_thai"
		}
	}

	// Sắp xếp theo NgayTaoRa
	orderBy := "ngay_tao_ra DESC"
	if sort == "views" {
		orderBy = "views DESC"
	}

	query.Count(&total)
	query.Order(orderBy).Offset(offset).Limit(limit).Find(&podcasts)
	c.JSON(http.StatusOK, gin.H{
		"data": podcasts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func SearchPodcast(c *gin.Context) {
	// Lấy các tham số từ query string
	search := c.Query("q")
	sortField := c.Query("sort")
	sortOrder := c.DefaultQuery("order", "desc") // Mặc định là DESC nếu không truyền
	trangThai := c.Query("trang_thai")           // "Bật", "Tắt", hoặc ""
	danhMucID := c.Query("danh_muc_id")          // Có thể lọc theo danh mục

	// Phân trang
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var podcasts []models.Podcast
	var total int64

	query := config.DB.Model(&models.Podcast{})

	// Lọc theo từ khóa tìm kiếm nếu có
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(tieu_de) LIKE ? OR LOWER(mo_ta) LIKE ? OR LOWER(the_tag) LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Lọc theo trạng thái nếu được truyền
	if trangThai != "" {
		query = query.Where("trang_thai = ?", trangThai)
	}

	// Lọc theo danh mục nếu được truyền
	if danhMucID != "" {
		query = query.Where("danh_muc_id = ?", danhMucID)
	}

	// Mapping trường sắp xếp hợp lệ
	switch sortField {
	case "ngay_tao_ra", "luot_xem", "tieu_de":
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}
		query = query.Order(sortField + " " + sortOrder)
	default:
		// Không sắp xếp nếu không hợp lệ
	}

	// Preload nếu có quan hệ liên kết
	query = query.Preload("TaiLieu").Preload("DanhMuc")

	// Tính tổng số podcast cho phân trang
	query.Count(&total)

	// Lấy danh sách podcast theo phân trang
	if err := query.Offset(offset).Limit(limit).Find(&podcasts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tìm kiếm podcast"})
		return
	}

	// Trả về kết quả và phân trang
	c.JSON(http.StatusOK, gin.H{
		"data": podcasts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func GetPodcastByID(c *gin.Context) {
	id := c.Param("id")
	var podcast models.Podcast

	// Xác định vai trò người dùng từ middleware
	role, _ := c.Get("vai_tro")

	query := config.DB.Model(&models.Podcast{})
	if role != "admin" {
		// User chỉ được xem podcast có trạng thái "Bật"
		query = query.Where("trang_thai = ?", "Bật")
	}

	// Tìm podcast theo ID (và theo trạng thái nếu không phải admin)
	if err := query.First(&podcast, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy podcast"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy thông tin podcast"})
		}
		return
	}

	// Tăng view count
	if err := config.DB.Model(&podcast).UpdateColumn("luot_xem", gorm.Expr("luot_xem + ?", 1)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tăng view count"})
		return
	}

	// Lấy các podcast liên quan cùng danh mục (chỉ lấy "Bật" để show ra cho user)
	var related []models.Podcast
	config.DB.
		Where("danh_muc_id = ? AND id != ? AND trang_thai = ?", podcast.DanhMucID, podcast.ID, "Bật").
		Order("ngay_tao_ra DESC").Limit(5).
		Find(&related)

	// Trả về dữ liệu
	c.JSON(http.StatusOK, gin.H{
		"data":    podcast,
		"suggest": related,
	})
}

// Lọc theo danh mục và sắp xếp
func GetFilteredPodcasts(c *gin.Context) {
	categoryID := c.Query("category_id") // Lọc theo danh mục (nếu có)
	sort := c.DefaultQuery("alphabetical", "date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var podcasts []models.Podcast
	query := config.DB.Model(&models.Podcast{}).
		Where("trang_thai = ?", "Bật")

	if categoryID != "" {
		query = query.Where("danh_muc_id = ?", categoryID)
	}

	// Sắp xếp
	orderBy := "ngay_tao_ra DESC"
	switch sort {
	case "az":
		orderBy = "tieu_de ASC"
	case "za":
		orderBy = "tieu_de DESC"
	}
	query = query.Order(orderBy)

	// Tổng số podcast
	var total int64
	query.Count(&total)

	// Lấy danh sách theo trang
	if err := query.Offset(offset).Limit(limit).Find(&podcasts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy danh sách podcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": podcasts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// Lọc theo danh mục và sắp xếp theo lượt xem cao nhất
func GetFilteredPodcastsByMostviewed(c *gin.Context) {
	categoryID := c.Query("category_id") // Lọc theo danh mục (nếu có)
	sort := c.DefaultQuery("most_viewed", "date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var podcasts []models.Podcast
	query := config.DB.Model(&models.Podcast{}).
		Where("trang_thai = ?", "Bật")

	if categoryID != "" {
		query = query.Where("danh_muc_id = ?", categoryID)
	}

	// Sắp xếp
	orderBy := "ngay_tao_ra DESC"
	if sort == "views" {
		orderBy = "luot_xem DESC"
	}
	query = query.Order(orderBy)

	// Tổng số podcast
	var total int64
	query.Count(&total)

	// Lấy danh sách theo trang
	if err := query.Offset(offset).Limit(limit).Find(&podcasts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy danh sách podcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": podcasts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// Lọc theo danh mục và sắp xếp theo lượt thời lượng
func GetFilteredPodcastsByDuration(c *gin.Context) {
	categoryID := c.Query("category_id") // Lọc theo danh mục (nếu có)
	sort := c.DefaultQuery("duration", "date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var podcasts []models.Podcast
	query := config.DB.Model(&models.Podcast{}).
		Where("trang_thai = ?", "Bật")

	if categoryID != "" {
		query = query.Where("danh_muc_id = ?", categoryID)
	}

	// Sắp xếp
	orderBy := "ngay_tao_ra DESC"
	switch sort {
	case "duration_asc":
		orderBy = "thoi_luong_giay ASC"
	case "duration_desc":
		orderBy = "thoi_luong_giay DESC"
	}
	query = query.Order(orderBy)

	// Tổng số podcast
	var total int64
	query.Count(&total)

	// Lấy danh sách theo trang
	if err := query.Offset(offset).Limit(limit).Find(&podcasts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy danh sách podcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": podcasts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// /Tạo podcast
func CreatePodcastWithUpload(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetString("user_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không có file đính kèm"})
		return
	}

	tieuDe := c.PostForm("tieu_de")
	danhMucID := c.PostForm("danh_muc_id")
	if tieuDe == "" || danhMucID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tiêu đề hoặc danh mục"})
		return
	}

	moTa := c.PostForm("mo_ta")
	hinhAnh := ""
	if hinhAnhFile, err := c.FormFile("hinh_anh_dai_dien"); err == nil {
		imageURL, err := utils.UploadImageToSupabase(hinhAnhFile, uuid.New().String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể upload hình ảnh", "details": err.Error()})
			return
		}
		hinhAnh = imageURL
	}

	theTag := c.PostForm("the_tag")
	voice := c.DefaultPostForm("voice", "vi-VN-Chirp3-HD-Puck")
	speakingRateStr := c.DefaultPostForm("speaking_rate", "1.0")
	rateValue, err := strconv.ParseFloat(speakingRateStr, 64)
	if err != nil || rateValue <= 0 {
		rateValue = 1.0
	}

	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header không hợp lệ"})
		return
	}
	token := parts[1]

	respData, err := services.CallUploadDocumentAPI(file, userID, token, voice, rateValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi gọi UploadDocument", "details": err.Error()})
		return
	}

	taiLieuRaw, ok := respData["tai_lieu"]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu tài liệu từ UploadDocument", "resp": respData})
		return
	}

	taiLieuMap, ok := taiLieuRaw.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dữ liệu tài liệu không đúng định dạng", "tai_lieu_raw": taiLieuRaw})
		return
	}

	audioURL, ok := respData["audio_url"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy audio URL từ UploadDocument"})
		return
	}

	taiLieuID, ok := taiLieuMap["id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy ID tài liệu"})
		return
	}

	durationFloat, err := services.GetMP3DurationFromURL(audioURL)
	totalSeconds := int(durationFloat)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tính thời lượng", "details": err.Error()})
		return
	}

	podcast := models.Podcast{
		ID:             uuid.New().String(),
		TailieuID:      taiLieuID,
		TieuDe:         tieuDe,
		MoTa:           moTa,
		DuongDanAudio:  audioURL,
		ThoiLuongGiay:  totalSeconds,
		HinhAnhDaiDien: hinhAnh,
		DanhMucID:      danhMucID,
		TrangThai:      "Tắt",
		NguoiTao:       userID,
		NgayXuatBan:    nil,
		TheTag:         theTag,
		LuotXem:        0,
	}

	if err := db.Create(&podcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo podcast", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo podcast thành công",
		"podcast": gin.H{
			"id":                podcast.ID,
			"tai_lieu_id":       podcast.TailieuID,
			"tieu_de":           podcast.TieuDe,
			"mo_ta":             podcast.MoTa,
			"duong_dan_audio":   podcast.DuongDanAudio,
			"thoi_luong_giay":   podcast.ThoiLuongGiay,
			"hinh_anh_dai_dien": podcast.HinhAnhDaiDien,
			"danh_muc_id":       podcast.DanhMucID,
			"trang_thai":        podcast.TrangThai,
			"nguoi_tao":         podcast.NguoiTao,
			"ngay_xuat_ban":     podcast.NgayXuatBan,
			"the_tag":           podcast.TheTag,
			"luot_xem":          podcast.LuotXem,
		},
		"thoi_luong_hienthi": FormatSecondsToHHMMSS(totalSeconds),
	})
}

func FormatSecondsToHHMMSS(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// Cập nhật podcast
func UpdatePodcast(c *gin.Context) {
	// Kiểm tra quyền admin
	if role, _ := c.Get("vai_tro"); role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện hành động này"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	podcastID := c.Param("id")

	var podcast models.Podcast
	if err := db.First(&podcast, "id = ?", podcastID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Podcast không tồn tại"})
		return
	}

	// Lấy dữ liệu từ form
	tieuDe := c.PostForm("tieu_de")
	moTa := c.PostForm("mo_ta")
	theTag := c.PostForm("the_tag")
	danhMucID := c.PostForm("danh_muc_id")
	trangThai := c.PostForm("trang_thai")

	// Cập nhật nếu có giá trị
	if tieuDe != "" {
		podcast.TieuDe = tieuDe
	}
	if moTa != "" {
		podcast.MoTa = moTa
	}
	if theTag != "" {
		podcast.TheTag = theTag
	}
	if danhMucID != "" {
		podcast.DanhMucID = danhMucID
	}
	if trangThai != "" {
		podcast.TrangThai = trangThai

		if trangThai == "Bật" {
			now := time.Now()
			podcast.NgayXuatBan = &now
		}
	}

	// Upload hình ảnh nếu có
	if hinhAnhFile, err := c.FormFile("hinh_anh_dai_dien"); err == nil {
		if imageURL, err := utils.UploadImageToSupabase(hinhAnhFile, uuid.New().String()); err == nil {
			podcast.HinhAnhDaiDien = imageURL
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể upload hình ảnh", "details": err.Error()})
			return
		}
	}

	// Lưu vào database
	if err := db.Save(&podcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật podcast", "details": err.Error()})
		return
	}

	// Load lại đầy đủ quan hệ
	if err := db.Preload("TaiLieu.NguoiDung").Preload("DanhMuc").First(&podcast, "id = ?", podcastID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể load dữ liệu podcast", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật podcast thành công",
		"podcast": podcast,
	})
}
