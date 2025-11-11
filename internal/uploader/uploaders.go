package uploader

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// --- Fungsi 1: Untuk Gambar (Pengaduan) ---
type ImgbbResponse struct {
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
	Success bool `json:"success"`
}
func UploadToImgbb(data []byte) (string, error) {
	apiKey := os.Getenv("IMGBB_API_KEY")
	if apiKey == "" {
		log.Fatal("[FATAL] Environment variable IMGBB_API_KEY is not set")
	}
	imgBase64 := base64.StdEncoding.EncodeToString(data)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("key", apiKey)
	writer.WriteField("image", imgBase64)
	writer.Close()
	req, err := http.NewRequest("POST", "https://api.imgbb.com/1/upload", body)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Failed to send request to imgbb: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Failed to read response: %v", err)
	}
	var imgbbResp ImgbbResponse
	if err := json.Unmarshal(respBody, &imgbbResp); err != nil {
		log.Printf("[WARN] Non-JSON response received from imgbb: %s", string(respBody))
		return "", fmt.Errorf("[ERROR] Failed to parse JSON: %v", err)
	}
	if !imgbbResp.Success {
		return "", fmt.Errorf("[ERROR] imgbb API returned an error response: %s", string(respBody))
	}
	log.Printf("[INFO] Image successfully uploaded to imgbb: %s", imgbbResp.Data.URL)
	return imgbbResp.Data.URL, nil
}


// --- FUNGSI BARU: Untuk File (PDF Surat) ---
func UploadFile(data []byte, fileName string) (string, error) {
	log.Printf("[INFO] Mengunggah file %s ke 0x0.st...", fileName)
	
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	writer.Close()
	req, err := http.NewRequest("POST", "http://0x0.st", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	url := strings.TrimSpace(string(respBody))
	if !strings.HasPrefix(url, "http") {
		return "", fmt.Errorf("gagal upload file, respons server: %s", url)
	}
	log.Printf("[INFO] File berhasil di-upload ke: %s", url)
	return url, nil
}
