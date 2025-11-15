package surat

import (
	"fmt"
	"log"
	"os"

	"github.com/sinavarasina/SAS-BOT/internal/uploader"
)

// UploadPDFToR2 membungkus logika upload R2
func UploadPDFToR2(pdfPath, pdfName string, r2Uploader *uploader.R2Uploader) (string, error) {
	log.Printf("[SURAT] Mengunggah %s ke Cloudflare R2...", pdfName)
	
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Printf("[SURAT-ERROR] Gagal membaca PDF untuk di-upload: %v", err)
		return "", err
	}

	fileURL, err := r2Uploader.UploadToR2(pdfData, pdfName) // Panggil uploader global
	if err != nil {
		log.Printf("[SURAT-ERROR] Gagal upload file ke R2: %v", err)
		return "Gagal Upload", err
	}
	
	log.Printf("[SURAT] Berhasil upload file ke R2: %s", fileURL)
	return fileURL, nil
}
