// file: internal/uploader/googledrive.go

package uploader

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveClient struct {
	Service  *drive.Service
	FolderID string
}

func InitDriveClient() (*DriveClient, error) {
	ctx := context.Background()
	credentialFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credentialFile), option.WithScopes(drive.DriveFileScope))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat Drive service: %w", err)
	}

	folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID")
	if folderID == "" {
		log.Fatal("Environment variable GOOGLE_DRIVE_FOLDER_ID tidak di-set")
	}

	return &DriveClient{
		Service:  srv,
		FolderID: folderID,
	}, nil
}

// UploadToDrive mengunggah file ke folder tertentu di Drive dan mengembalikan ID file.
func (c *DriveClient) UploadToDrive(data []byte, fileName string) (string, error) {
	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{c.FolderID},
	}

	file, err := c.Service.Files.Create(fileMetadata).Media(bytes.NewReader(data)).Do()
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah file ke Drive: %w", err)
	}

	// Setelah upload, buat file menjadi publik (bisa dilihat siapa saja dengan link)
	_, err = c.Service.Permissions.Create(file.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		return "", fmt.Errorf("gagal mengatur izin file: %w", err)
	}

	log.Printf("File berhasil diunggah ke Drive dengan ID: %s", file.Id)
	return file.Id, nil
}
