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
	ParentID string // Ini adalah ID folder induk "SAS-BOT Arsip"
}

func InitDriveClient() (*DriveClient, error) {
	ctx := context.Background()
	credentialFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// Pastikan Scope-nya adalah DriveFileScope
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credentialFile), option.WithScopes(drive.DriveFileScope))
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to create Drive service: %w", err)
	}

	// Baca ID Folder Induk dari .env
	parentID := os.Getenv("GOOGLE_DRIVE_PARENT_ID")
	if parentID == "" {
		log.Fatal("[FATAL] Environment variable GOOGLE_DRIVE_PARENT_ID is not set")
	}

	return &DriveClient{
		Service:  srv,
		ParentID: parentID,
	}, nil
}

// GetOrCreateFolder adalah fungsi kunci: mencari folder, jika tidak ada, membuatnya.
func (c *DriveClient) GetOrCreateFolder(parentID, folderName string) (string, error) {
	// 1. Buat query untuk mencari folder dengan nama dan parent yang sama
	query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName, parentID)
	
	r, err := c.Service.Files.List().Q(query).PageSize(1).Fields("files(id)").Do()
	if err != nil {
		return "", fmt.Errorf("gagal mencari folder: %w", err)
	}

	// 2. Jika ditemukan, kembalikan ID-nya
	if len(r.Files) > 0 {
		log.Printf("[DRIVE] Folder '%s' ditemukan, ID: %s", folderName, r.Files[0].Id)
		return r.Files[0].Id, nil
	}

	// 3. Jika tidak ditemukan, buat folder baru
	log.Printf("[DRIVE] Folder '%s' tidak ditemukan, membuat baru...", folderName)
	folderMetadata := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	file, err := c.Service.Files.Create(folderMetadata).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder: %w", err)
	}

	log.Printf("[DRIVE] Folder baru '%s' dibuat, ID: %s", folderName, file.Id)
	return file.Id, nil
}


// UploadToDrive sekarang menerima folderID tujuan
func (c *DriveClient) UploadToDrive(data []byte, fileName string, folderID string) (string, error) {
	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{folderID}, // Upload ke folderID spesifik
	}

	file, err := c.Service.Files.Create(fileMetadata).Media(bytes.NewReader(data)).Do()
	if err != nil {
		return "", fmt.Errorf("[ERROR] Failed to upload file to Google Drive: %v", err)
	}

	// Jadikan file publik (bisa dilihat siapa saja dengan link)
	_, err = c.Service.Permissions.Create(file.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		return "", fmt.Errorf("[ERROR] Failed to set file permissions: %v", err)
	}

	log.Printf("[INFO] File successfully uploaded to Google Drive with ID: %s", file.Id)
	return file.Id, nil
}
