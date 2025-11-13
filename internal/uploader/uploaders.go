package uploader

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws" 
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func UploadToR2(data []byte, fileName string) (string, error) {
    accountID := os.Getenv("R2_ACCOUNT_ID")
    accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
    secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
    bucketName := os.Getenv("R2_BUCKET_NAME")
    publicURL := os.Getenv("R2_PUBLIC_URL")

    if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" || publicURL == "" {
        return "", fmt.Errorf("kredensial R2 tidak di-set di .env")
    }

    log.Printf("[R2] Mengunggah %s ke bucket %s...", fileName, bucketName)

    endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

    // Resolver endpoint R2
    resolver := aws.EndpointResolverWithOptionsFunc(
        func(service, region string, options ...interface{}) (aws.Endpoint, error) {
            return aws.Endpoint{
                URL:           endpoint,
                SigningRegion: "auto",
            }, nil
        },
    )

    // Konfigurasi AWS SDK
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithEndpointResolverWithOptions(resolver),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
        config.WithRegion("auto"),
    )
    if err != nil {
        return "", fmt.Errorf("gagal memuat konfigurasi AWS/R2: %w", err)
    }

    client := s3.NewFromConfig(cfg)

    // Upload file
    _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: &bucketName,
        Key:    &fileName,
        Body:   bytes.NewReader(data),
        ACL:    "public-read",
    })
    if err != nil {
        return "", fmt.Errorf("gagal mengunggah file ke R2: %w", err)
    }

    finalURL := fmt.Sprintf("%s/%s", publicURL, fileName)
    log.Printf("[R2] Berhasil upload: %s", finalURL)

    return finalURL, nil
}
