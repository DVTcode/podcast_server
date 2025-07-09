package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/DVTcode/podcast_server/models"
)

type RegisterInput struct {
	Email   string `json:"email" binding:"required,email"`
	MatKhau string `json:"mat_khau" binding:"required,min=6"`
	HoTen   string `json:"ho_ten" binding:"required"`
}

func Register(c *gin.Context, db *gorm.DB) {
	var input RegisterInput

	// Parse JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check email đã tồn tại chưa
	var existing models.NguoiDung
	if err := db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email đã được sử dụng"})
		return
	}
	if input.MatKhau == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mật khẩu không được để trống"})
		return
	}
	if len(input.MatKhau) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mật khẩu phải có ít nhất 6 ký tự"})
		return
	}
	if input.HoTen == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Họ tên không được để trống"})
		return
	}
	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.MatKhau), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể mã hoá mật khẩu"})
		return
	}

	// Tạo user mới
	newUser := models.NguoiDung{
		ID:       uuid.New().String(),
		Email:    input.Email,
		MatKhau:  string(hashedPassword),
		HoTen:    input.HoTen,
		VaiTro:   "user",
		KichHoat: true,
	}

	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo người dùng"})
		return
	}

	// Ẩn mật khẩu khi trả về
	newUser.MatKhau = ""
	c.JSON(http.StatusCreated, newUser)
}
