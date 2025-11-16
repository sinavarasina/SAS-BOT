package uploader

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadToR2 mengunggah file (PDF) ke bucket R2
func (u *R2Uploader) UploadToR2(data []byte, fileName string) (string, error) {
	if u.AccountID == "" || u.AccessKeyID == "" || u.SecretAccessKey == "" || u.BucketName == "" || u.PublicURL == "" {
		return "", fmt.Errorf("kredensial R2 tidak di-set")
	}

	log.Printf("[R2] Mengunggah %s ke bucket %s...", fileName, u.BucketName)

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", u.AccountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(u.AccessKeyID, u.SecretAccessKey, "")),
	)
	if err != nil {
		return "", fmt.Errorf("gagal memuat konfigurasi AWS/R2: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &u.BucketName,
		Key:    &fileName,
		Body:   bytes.NewReader(data),
		ACL:    "public-read",
	})
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah file ke R2: %w", err)
	}

	finalURL := fmt.Sprintf("%s/%s", u.PublicURL, fileName)
	log.Printf("[R2] Berhasil upload file ke: %s", finalURL)
	return finalURL, nil
}
