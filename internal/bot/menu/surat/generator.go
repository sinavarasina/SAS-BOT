package surat

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
)

// GenerateAsync menjalankan kompilasi LaTeX di background
func GenerateAsync(
	template JenisSurat,
	data map[string]string,
	tempDir string,
	jid string,
	client *whatsmeow.Client,
	pdfName string,
	uniqueID string,
	sheetsClient *sheets.SheetsClient,
	r2Uploader *uploader.R2Uploader,
) (string, error) {

	src := filepath.Join("templates", string(template))
	texBytes, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("gagal membaca template: %w", err)
	}

	relLogoPath := filepath.Join("..", "templates", "logo", "logo.jpg")
	data["LOGOPATH"] = filepath.ToSlash(relLogoPath)

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
		cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-output-directory", absTemp, filepath.Base(texPath))
		cmd.Dir = absTemp
		output, err := cmd.CombinedOutput()

		if _, statErr := os.Stat(pdfPath); os.IsNotExist(statErr) {
			log.Printf("[Surat-Error] pdflatex gagal compile: %v\n%s", err, output)
			_ = SendMessage(client, jid, fmt.Sprintf("Surat %s gagal dikompilasi.", template))
			return
		}
		if err != nil {
			log.Printf("[SURAT-WARN] pdflatex selesai dengan peringatan (error): %v\n%s", err, output)
		}

		log.Printf("[SURAT] PDF selesai dibuat: %s", pdfPath)

		// 1. Kirim ke Pengguna
		err = SendFile(client, jid, pdfPath,
			fmt.Sprintf("Surat *%s* Anda telah selesai dibuat.\nNomor Unik Surat: *%s*", NamaSuratmap[template], uniqueID))
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal mengirim file PDF ke user: %v", err)
			_ = SendMessage(client, jid, "File berhasil dibuat tapi gagal dikirim. Silakan hubungi admin.")
		}

		// 2. Kirim ke Perangkat Desa
		nomorKades := os.Getenv("NOMOR_PERANGKAT_DESA")
		if nomorKades != "" {
			kadesJID := fmt.Sprintf("%s@s.whatsapp.net", nomorKades)
			log.Printf("[SURAT] Mengirim salinan PDF ke Perangkat Desa (%s)", kadesJID)
			_ = SendFile(client, kadesJID, pdfPath,
				fmt.Sprintf("Laporan Surat Baru Dibuat:\nJenis: %s\nID Unik: %s\nAtas Nama: %s", NamaSuratmap[template], uniqueID, data["NAMA"]))
		}

		// 3. Upload ke R2
		fileURL, err := UploadPDFToR2(pdfPath, pdfName, r2Uploader)
		if err != nil {
			fileURL = "Gagal Upload"
		}
		
		// 4. Log ke Sheet
		tgl := time.Now().Format("02-01-2006")
		sheetsClient.AppendSuratLog(string(template), data["NAMA"], uniqueID, tgl, "Belum Diproses", fileURL)

		// 5. Hapus file sementara
		os.Remove(texPath)
		os.Remove(pdfPath)
		os.Remove(strings.Replace(texPath, ".tex", ".log", 1))
		os.Remove(strings.Replace(texPath, ".tex", ".aux", 1))
		log.Printf("[SURAT] File sementara untuk %s telah dibersihkan.", pdfName)

	}()

	return texPath, nil
}
