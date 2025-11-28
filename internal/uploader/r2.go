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

func (u *R2Uploader) UploadToR2(data []byte, fileName string) (string, error) {
	if u.AccountID == "" || u.AccessKeyID == "" || u.SecretAccessKey == "" || u.BucketName == "" || u.PublicURL == "" {
		return "", fmt.Errorf("kredensial R2 tidak lengkap di types.go/env")
	}

	log.Printf("[R2] Mengunggah %s ke bucket %s...", fileName, u.BucketName)

	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", u.AccountID)

	// Konfigurasi Custom Endpoint Resolver
	customEndpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               r2Endpoint,
			HostnameImmutable: true,
			SigningRegion:     "auto",
		}, nil
	})

	// Muat konfigurasi AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(u.AccessKeyID, u.SecretAccessKey, "")),
		config.WithRegion("auto"),
		config.WithEndpointResolverWithOptions(customEndpointResolver),
	)

	if err != nil {
		return "", fmt.Errorf("gagal memuat konfigurasi AWS/R2: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(u.BucketName),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/pdf"), 
	})

	if err != nil {
		return "", fmt.Errorf("gagal mengunggah file ke R2: %w", err)
	}

	// Construct Public URL
	finalURL := fmt.Sprintf("%s/%s", u.PublicURL, fileName)
	log.Printf("[R2] Berhasil upload file ke: %s", finalURL)
	return finalURL, nil
}
