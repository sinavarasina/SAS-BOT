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

func escapeLatex(text string) string {
	replacer := strings.NewReplacer(
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
		"\\", "\\textbackslash{}",
	)
	return replacer.Replace(text)
}

func fillTemplate(content string, data map[string]string) string {
	for key, value := range data {
		cleanValue := value
		
		// Jangan escape path logo karena itu path file sistem
		if key != "LOGOPATH" {
			cleanValue = escapeLatex(value)
		}

		// Handle khusus jika ada key ALASANPERLU (spasi tambahan)
		if key == "ALASANPERLU" {
			content = strings.ReplaceAll(content, "{{"+key+"}}", " "+cleanValue+" ")
		} else {
			content = strings.ReplaceAll(content, "{{"+key+"}}", cleanValue)
		}
	}
	return content
}

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

	// --- LOGIC TAMBAHAN (SOLUSI CRASH) ---
	// 1. Setup Logo Path
	relLogoPath := filepath.Join("..", "templates", "logo", "logo.jpg")
	data["LOGOPATH"] = filepath.ToSlash(relLogoPath)
	
	// 2. Setup Nama Penandatangan (Kades)
	signer := os.Getenv("SIGNER_NAME")
	if signer == "" {
		signer = "........................"
	}
	data["SIGNER.NAME"] = signer

	// 3. Setup Tanggal Surat & Tahun (INJEKSI OTOMATIS)
	// Tanpa ini, {{TANGGAL_SURAT}} akan tersisa di tex dan underscore-nya bikin crash
	now := time.Now()
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	monthStr := months[now.Month()]
	
	// Format: 29 Desember 2025
	data["TANGGAL_SURAT"] = fmt.Sprintf("%d %s %d", now.Day(), monthStr, now.Year())
	data["TAHUN"] = fmt.Sprintf("%d", now.Year())
	// -------------------------------------

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
		
		cmd := exec.Command("pdflatex",
			"-interaction=nonstopmode",
			"-output-directory", absTemp,
			filepath.Base(texPath),
		)
		cmd.Dir = absTemp 

		output, err := cmd.CombinedOutput()

		// Cek apakah PDF benar-benar terbentuk
		if _, statErr := os.Stat(pdfPath); os.IsNotExist(statErr) {
			log.Printf("[Surat-Error] pdflatex gagal compile: %v\nOutput LaTeX:\n%s", err, output)
			_ = SendMessage(client, jid, fmt.Sprintf("Maaf, Surat %s gagal dibuat karena kesalahan format.", template))
			return
		}
		
		if err != nil {
			// Jika error tapi PDF ada, biasanya warning minor.
			log.Printf("[SURAT-WARN] pdflatex warning: %v", err)
		}

		log.Printf("[SURAT] PDF selesai dibuat: %s", pdfPath)

		// 1. Kirim ke Pengguna
		caption := fmt.Sprintf("✅ Surat *%s* Anda telah selesai.\n🆔 ID: *%s*", NamaSuratmap[template], uniqueID)
		err = SendFile(client, jid, pdfPath, caption)
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal kirim WA ke user: %v", err)
			_ = SendMessage(client, jid, "File PDF sudah jadi, tapi gagal dikirim via WA. Silakan cek status nanti.")
		}

		// 2. Kirim ke Perangkat Desa
		nomorKades := os.Getenv("NOMOR_PERANGKAT_DESA")
		if nomorKades != "" {
			kadesJID := fmt.Sprintf("%s@s.whatsapp.net", nomorKades)
			// Gunakan NAMA_PEMOHON jika NAMA kosong (biasanya key di map berbeda)
			namaPemohon := data["NAMA"]
			if namaPemohon == "" {
				namaPemohon = data["NAMA_PEMOHON"]
			}
			
			_ = SendFile(client, kadesJID, pdfPath,
				fmt.Sprintf("Laporan Surat Baru Dibuat:\nJenis: %s\nID Unik: %s\nAtas Nama: %s", NamaSuratmap[template], uniqueID, namaPemohon))
		}

		// 3. Upload ke R2
		fileURL, err := UploadPDFToR2(pdfPath, pdfName, r2Uploader)
		if err != nil {
			log.Printf("[R2-ERROR] Gagal upload: %v", err)
			fileURL = "Gagal Upload"
		}

		// 4. Log ke Sheet
		// Gunakan NAMA_PEMOHON jika NAMA kosong
		namaLog := data["NAMA"]
		if namaLog == "" {
			namaLog = data["NAMA_PEMOHON"]
		}

		tglLog := time.Now().Format("02-01-2006")
		sheetsClient.AppendSuratLog(string(template), namaLog, uniqueID, tglLog, "Belum Diproses", fileURL)

		// 5. Hapus file sementara
		time.Sleep(2 * time.Second)
		os.Remove(texPath)
		os.Remove(pdfPath)
		os.Remove(strings.Replace(texPath, ".tex", ".log", 1))
		os.Remove(strings.Replace(texPath, ".tex", ".aux", 1))
		log.Printf("[SURAT] File sementara untuk %s telah dibersihkan.", pdfName)

	}()

	return texPath, nil
}
