package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func CallUploadDocumentAPI(file *multipart.FileHeader, userID string, token string, voice string, speakingRate float64) (map[string]interface{}, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return nil, err
	}

	fileContent, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileContent.Close()
	if _, err = io.Copy(fw, fileContent); err != nil {
		return nil, err
	}

	_ = writer.WriteField("voice", voice)
	_ = writer.WriteField("speaking_rate", fmt.Sprintf("%f", speakingRate))
	writer.Close()

	baseURL := os.Getenv("API_BASE_URL")
	println("Base URL:", baseURL)

	req, err := http.NewRequest("POST", "https://podcastserver-production.up.railway.app/api/admin/documents/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("user_id", userID)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, err
	}
	return result, nil
}
