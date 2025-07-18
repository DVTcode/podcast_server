package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/DVTcode/podcast_server/config"
	"github.com/DVTcode/podcast_server/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetPodcast(c *gin.Context) {
	var podcasts []models.Podcast
	var total int64

	// Phân trang
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Tìm kiếm & lọc
	search := c.Query("search")
	status := c.Query("status")            // "true"/"false"
	categoryID := c.Query("category")      // lọc theo danh mục
	sort := c.DefaultQuery("sort", "date") // sắp xếp theo ngày tạo hoặc tên
	query := config.DB.Model(&models.Podcast{})
	// Lấy role từ context (middleware đã set)
	role, _ := c.Get("vai_tro")
	if role != "admin" {
		query = query.Where("kich_hoat = ?", true) // chỉ lấy podcast đã kích hoạt
	}
	if search != "" {
		query = query.Where("LOWER(tieu_de) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" && role == "admin" {
		switch status {
		case "true":
			query = query.Where("kich_hoat = ?", true)
		case "false":
			query = query.Where("kich_hoat = ?", false)
		}
	}

	orderBy := "created_at DESC" // mặc định sắp xếp theo ngày tạo mới nhất
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
	search := c.Query("q")
	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu từ khoá tìm kiếm"})
		return
	}

	var podcasts []models.Podcast
	query := config.DB.Model(&models.Podcast{}).
		Where("LOWER(tieu_de) LIKE ?", "%"+strings.ToLower(search)+"%").
		Or("LOWER(mo_ta) LIKE ?", "%"+strings.ToLower(search)+"%").
		Where("kich_hoat = ?", true) // chỉ lấy podcast đã kích hoạt

	if err := query.Find(&podcasts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tìm kiếm podcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": podcasts})
}

func GetPodcastByID(c *gin.Context) {
	id := c.Param("id")
	var podcast models.Podcast

	// Lấy podcast theo id
	if err := config.DB.First(&podcast, "id = ? AND kich_hoat = ?", id, true).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy podcast"})
		return
	}

	// Tăng view count
	config.DB.Model(&podcast).UpdateColumn("views", gorm.Expr("views + ?", 1))

	// Gợi ý podcast liên quan (cùng category, loại trừ chính nó, lấy 5 cái mới nhất)
	var related []models.Podcast
	config.DB.Where("danh_muc_id = ? AND id != ? AND kich_hoat = ?", podcast.DanhMucID, podcast.ID, true).
		Order("created_at DESC").Limit(5).Find(&related)

	c.JSON(http.StatusOK, gin.H{
		"data":    podcast,
		"suggest": related,
	})
}
