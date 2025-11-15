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
)

type ImgbbResponse struct {
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
	Success bool `json:"success"`
}

// UploadToImgbb sekarang adalah method dari struct
func (u *ImgbbUploader) UploadToImgbb(data []byte) (string, error) {
	if u.APIKey == "" {
		return "", fmt.Errorf("[FATAL] IMGBB_API_KEY tidak di-set")
	}
	imgBase64 := base64.StdEncoding.EncodeToString(data)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("key", u.APIKey)
	writer.WriteField("image", imgBase64)
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.imgbb.com/1/upload", body)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Gagal membuat request imgbb: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Gagal mengirim request ke imgbb: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("[ERROR] Gagal membaca respons imgbb: %v", err)
	}

	var imgbbResp ImgbbResponse
	if err := json.Unmarshal(respBody, &imgbbResp); err != nil {
		return "", fmt.Errorf("[ERROR] Gagal parse JSON imgbb: %v", err)
	}
	if !imgbbResp.Success {
		return "", fmt.Errorf("[ERROR] imgbb API mengembalikan error: %s", string(respBody))
	}

	log.Printf("[INFO] Image berhasil di-upload ke imgbb: %s", imgbbResp.Data.URL)
	return imgbbResp.Data.URL, nil
}
