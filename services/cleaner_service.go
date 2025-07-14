package services

import (
	"regexp"
	"strings"
)

// CleanText xử lý nội dung để chuẩn bị cho AI:
// - loại bỏ ký tự đặc biệt
// - chuẩn hóa khoảng trắng
// - fix encoding (nếu có), loại bỏ BOM, vv.
func CleanText(input string) string {
	// Loại bỏ BOM nếu có
	input = strings.TrimPrefix(input, "\uFEFF")

	// Loại bỏ các ký tự không phải chữ/số/câu thông thường
	reg := regexp.MustCompile(`[^\p{L}\p{N}\p{P}\p{Z}\n\r\t]+`)
	clean := reg.ReplaceAllString(input, "")

	// Chuẩn hóa các dấu xuống dòng về "\n"
	clean = strings.ReplaceAll(clean, "\r\n", "\n")
	clean = strings.ReplaceAll(clean, "\r", "\n")

	// Chuẩn hóa khoảng trắng
	clean = strings.Join(strings.Fields(clean), " ")

	// Loại bỏ khoảng trắng dư
	return strings.TrimSpace(clean)
}
