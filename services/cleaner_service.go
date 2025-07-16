package services

import (
	"regexp"
	"strings"

	"github.com/DVTcode/podcast_server/utils"
)

// PreCleanText xử lý thô: loại mục lục, số trang, code, khoảng trắng
func PreCleanText(text string) string {
	cleaned := text

	// Xoá các dòng chứa "Mục lục" hoặc "Table of Contents"
	reTOC := regexp.MustCompile(`(?i)^(.*mục lục.*|.*table of contents.*)$`)
	cleaned = reTOC.ReplaceAllString(cleaned, "")

	// Xoá các dòng chứa "Trang X" hoặc "Page X"
	rePageNumber := regexp.MustCompile(`(?i)^.*(trang|page)[^\d]*\d+.*$`)
	cleaned = rePageNumber.ReplaceAllString(cleaned, "")

	// Xoá dòng chỉ có số, ký tự đặc biệt hoặc khoảng trắng
	reSpecialLines := regexp.MustCompile(`^[\s\W\d]*$`)
	cleaned = reSpecialLines.ReplaceAllString(cleaned, "")

	// Xoá dòng có chứa code hoặc từ khoá lập trình
	reCode := regexp.MustCompile(`(?i)^.*(const |function |class |<[^>]+>).*?$`)
	cleaned = reCode.ReplaceAllString(cleaned, "")

	// Xoá nhiều dòng trống liên tiếp
	reMultiNewLine := regexp.MustCompile(`\n{2,}`)
	cleaned = reMultiNewLine.ReplaceAllString(cleaned, "\n")

	return strings.TrimSpace(cleaned)
}

// CleanWithGemini sử dụng Gemini để làm sạch sâu, chuẩn hoá văn bản
func CleanWithGemini(text string) (string, error) {
	prompt := `Bạn là công cụ xử lý văn bản trích xuất từ tài liệu.
	Hãy xử lý văn bản sau với yêu cầu:
	- Xoá phần mục lục, các dòng chứa số trang, tiêu đề lặp lại
	- Xoá code, ví dụ mã lệnh, hoặc các ký hiệu kỹ thuật
	- Làm gọn văn bản: không có dòng trống thừa, không có ký tự lạ
	- Ngắt đoạn hợp lý, dễ đọc, phù hợp để chuyển thành nội dung podcast
	- Giữ nguyên nội dung, không thêm bớt, không giải thích

	Văn bản cần làm sạch:`

	fullPrompt := prompt + "\n\n" + text

	return utils.GeminiGenerateText(fullPrompt)
}

// CleanTextPipeline là pipeline chính: Regex + Gemini
func CleanTextPipeline(rawText string) (string, error) {
	preCleaned := PreCleanText(rawText)
	finalCleaned, err := CleanWithGemini(preCleaned)
	if err != nil {
		return "", err
	}
	return finalCleaned, nil
}
