package services

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return "", fmt.Errorf("không thể khởi tạo Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	prompt := `Bạn là công cụ xử lý văn bản trích xuất từ tài liệu.
Hãy xử lý văn bản sau với yêu cầu:
- Xoá phần mục lục, các dòng chứa số trang, tiêu đề lặp lại
- Xoá code, ví dụ mã lệnh, hoặc các ký hiệu kỹ thuật
- Làm gọn văn bản: không có dòng trống thừa, không có ký tự lạ
- Ngắt đoạn hợp lý, dễ đọc, phù hợp để chuyển thành nội dung podcast
- Giữ nguyên nội dung, không thêm bớt, không giải thích

Văn bản cần làm sạch:`

	fullPrompt := prompt + "\n\n" + text

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("Gemini xử lý lỗi: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Không nhận được kết quả từ Gemini")
	}

	return strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])), nil
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
