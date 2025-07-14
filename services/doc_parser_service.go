package services

import (
	"fmt"
	"io"
	"os"

	"mime/multipart"

	"github.com/nguyenthenguyen/docx"
)

// ExtractTextFromDOCX là parser thay thế miễn phí cho file DOCX
func ExtractTextFromDOCX(fileHeader *multipart.FileHeader) (string, error) {
	// Tạo file tạm
	tempFilePath := fmt.Sprintf("./temp/%s", fileHeader.Filename)
	if err := os.MkdirAll("./temp", 0755); err != nil {
		return "", err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	io.Copy(dst, src)

	// Mở file bằng docx lib
	r, err := docx.ReadDocxFile(tempFilePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	doc := r.Editable()
	return doc.GetContent(), nil
}
