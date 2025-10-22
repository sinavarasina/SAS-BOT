// file: internal/uploader/imgbb.go

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
)

// ImgbbResponse adalah struct untuk menangkap respons JSON
type ImgbbResponse struct {
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
	Success bool `json:"success"`
}

// UploadToImgbb mengunggah data gambar dan mengembalikan URL publiknya
func UploadToImgbb(data []byte) (string, error) {
	apiKey := os.Getenv("IMGBB_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable IMGBB_API_KEY tidak di-set")
	}

	// Data gambar perlu di-encode ke Base64 untuk dikirim via form
	imgBase64 := base64.StdEncoding.EncodeToString(data)

	// Persiapkan form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("key", apiKey)
	writer.WriteField("image", imgBase64)
	writer.Close()

	// Buat request
	req, err := http.NewRequest("POST", "https://api.imgbb.com/1/upload", body)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Kirim request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengirim request ke imgbb: %w", err)
	}
	defer resp.Body.Close()

	// Baca respons
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respons: %w", err)
	}

	// Parse JSON
	var imgbbResp ImgbbResponse
	if err := json.Unmarshal(respBody, &imgbbResp); err != nil {
		log.Printf("Respons non-JSON dari imgbb: %s", string(respBody))
		return "", fmt.Errorf("gagal parse JSON: %w", err)
	}

	if !imgbbResp.Success {
		return "", fmt.Errorf("imgbb API mengembalikan error: %s", string(respBody))
	}

	log.Printf("Gambar berhasil diunggah ke imgbb: %s", imgbbResp.Data.URL)
	return imgbbResp.Data.URL, nil
}
