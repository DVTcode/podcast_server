package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	storage "github.com/supabase-community/storage-go"
)

func UploadFileToSupabase(fileHeader *multipart.FileHeader, fileID string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	storageClient := storage.NewClient(supabaseURL+"/storage/v1", supabaseKey, nil)

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	objectPath := fmt.Sprintf("uploads/uploads/%s%s", fileID, ext)

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return "", err
	}

	contentType := fileHeader.Header.Get("Content-Type")
	options := storage.FileOptions{
		ContentType: &contentType,
	}

	_, err = storageClient.UploadFile("uploads", objectPath, &buf, options)
	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s", supabaseURL, objectPath)
	return publicURL, nil
}
