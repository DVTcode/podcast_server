package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// Định nghĩa loại input
type InputType string

const (
	InputText  InputType = "text"
	InputTXT   InputType = "txt"
	InputDOCX  InputType = "docx"
	InputPDF   InputType = "pdf"
	InputAudio InputType = "audio" // Dành cho bước sau (nếu cần tích hợp Speech-to-Text)
)

// Struct đại diện cho nguồn input
type InputSource struct {
	Type       InputType
	FileHeader *multipart.FileHeader // Nếu là file (txt, docx, pdf, audio)
	Text       string                // Nếu người dùng nhập tay
}

// Hàm xử lý input thành plain text
func NormalizeInput(file *multipart.FileHeader) (string, error) {
	// mở file từ fileHeader
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".txt":
		return ExtractTextFromTXT(file)
	case ".pdf":
		return ExtractTextFromPDF(f)
	case ".doc", ".docx":
		return ExtractTextFromDOC(file)
	default:
		return "", errors.New("Định dạng file không được hỗ trợ")
	}
}
