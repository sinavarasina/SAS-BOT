package surat

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	// "github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
)

func fillTemplate(content string, data map[string]string) string {
	for key, value := range data {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
}

func GenerateAsync(template JenisSurat, data map[string]string, tempDir string, jid string, client *whatsmeow.Client, pdfName string, uniqueID string, sheetsClient *sheets.SheetsClient) (string, error) {
	src := filepath.Join("templates", string(template))
	texBytes, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("gagal membaca template: %w", err)
	}

	relLogoPath := filepath.Join("..", "templates", "logo", "logo.jpg")
	data["LOGOPATH"] = filepath.ToSlash(relLogoPath)
	signer := os.Getenv("SIGNER_NAME")
	if signer == "" {
		signer = "........................"
	}
	data["SIGNER.NAME"] = signer

	filled := fillTemplate(string(texBytes), data)
	texPath := filepath.Join(tempDir, strings.Replace(pdfName, ".pdf", ".tex", 1))
	pdfPath := filepath.Join(tempDir, pdfName)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder %s: %w", tempDir, err)
	}
	if err := os.WriteFile(texPath, []byte(filled), 0644); err != nil {
		return "", fmt.Errorf("gagal menulis file %s: %w", texPath, err)
	}

	go func() {
		absTemp, _ := filepath.Abs(tempDir)
		// pdfPath := strings.Replace(texPath, ".tex", ".pdf", 1)

		cmd := exec.Command("pdflatex",
			"-interaction=nonstopmode",
			"-output-directory", absTemp,
			filepath.Base(texPath),
		)
		cmd.Dir = absTemp // Direktori kerja adalah 'temp/'

		output, err := cmd.CombinedOutput()

		// Cek apakah file PDF benar-benar TIDAK ada
		if _, statErr := os.Stat(pdfPath); os.IsNotExist(statErr) {
			// Jika file PDF TIDAK ADA, ini adalah kegagalan total.
			log.Printf("[Surat-Error] pdflatex gagal compile: %v\n%s", err, output)
			_ = SendMessage(client, jid, fmt.Sprintf("Surat %s gagal dikompilasi. File PDF tidak dapat dibuat.", template))
			return
		}

		if err != nil {
			// Log error-nya sebagai peringatan, tapi jangan hentikan proses
			log.Printf("[SURAT-WARN] pdflatex selesai dengan peringatan (error): %v\n%s", err, output)
		}

		log.Printf("[SURAT] PDF selesai dibuat: %s", pdfPath)

		// 1. Kirim ke Pengguna
		err = SendFile(client, jid, pdfPath,
			fmt.Sprintf("Surat *%s* Anda telah selesai dibuat.\nNomor Unik Surat: *%s*", NamaSuratmap[template], uniqueID))
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal mengirim file PDF ke user: %v", err)
			_ = SendMessage(client, jid,
				"File berhasil dibuat tapi gagal dikirim. Silakan hubungi admin.")
		}

		nomorKades := os.Getenv("NOMOR_PERANGKAT_DESA")
		if nomorKades != "" {
			kadesJID := fmt.Sprintf("%s@s.whatsapp.net", nomorKades)
			log.Printf("[SURAT] Mengirim salinan PDF ke Perangkat Desa (%s)", kadesJID)
			_ = SendFile(client, kadesJID, pdfPath,
				fmt.Sprintf("Laporan Surat Baru Dibuat:\nJenis: %s\nID Unik: %s\nAtas Nama: %s", NamaSuratmap[template], uniqueID, data["NAMA"]))
		}

		log.Printf("[SURAT] Mengunggah %s ke cloud...", pdfName)
		pdfData, err := os.ReadFile(pdfPath)
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal membaca PDF untuk di-upload: %v", err)
			return
		}

		fileURL, err := uploader.UploadFile(pdfData, pdfName)
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal upload file: %v", err)
			fileURL = "Gagal Upload"
		} else {
			log.Printf("[SURAT] Berhasil upload file: %s", fileURL)
		}

		tgl := time.Now().Format("02-01-2006")
		sheetsClient.AppendSuratLog(string(template), data["NAMA"], uniqueID, tgl, "Belum Diproses", fileURL)
		// (Opsional tapi disarankan) Hapus file sementara setelah dikirim
		// os.Remove(texPath)
		// os.Remove(pdfPath)
		// os.Remove(strings.Replace(texPath, ".tex", ".log", 1))
		// os.Remove(strings.Replace(texPath, ".tex", ".aux", 1))

	}()

	return texPath, nil
}
