package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/unidoc/unioffice/document"
)

func ExtractTextFromDOC(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("không mở được file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return "", fmt.Errorf("lỗi đọc file: %w", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	doc, err := document.Read(reader, int64(reader.Len()))
	if err != nil {
		return "", fmt.Errorf("không đọc được file DOCX: %w", err)
	}

	var builder strings.Builder

	// Trích xuất đoạn văn
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			builder.WriteString(run.Text())
		}
		builder.WriteString("\n")
	}

	// Trích xuất bảng
	for _, tbl := range doc.Tables() {
		for _, row := range tbl.Rows() {
			for _, cell := range row.Cells() {
				for _, para := range cell.Paragraphs() {
					for _, run := range para.Runs() {
						builder.WriteString(run.Text() + "\t")
					}
				}
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
