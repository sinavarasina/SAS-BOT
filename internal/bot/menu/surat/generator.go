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

// escapeLatex mencegah error akibat karakter spesial (seperti _ $ & %)
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

// fillTemplate mengganti placeholder {{KEY}} dengan data
func fillTemplate(content string, data map[string]string) string {
	for key, value := range data {
		cleanValue := value
		
		// Jangan escape path logo karena itu path file sistem
		if key != "LOGOPATH" {
			cleanValue = escapeLatex(value)
		}

		// Handle khusus spasi untuk field paragraf panjang
		if key == "ALASANPERLU" || key == "KEPERLUAN" {
			content = strings.ReplaceAll(content, "{{"+key+"}}", " "+cleanValue+" ")
		} else {
			content = strings.ReplaceAll(content, "{{"+key+"}}", cleanValue)
		}
	}
	return content
}

// GenerateAsync menjalankan kompilasi LaTeX, Pengiriman WA, Upload R2, dan Logging Sheets
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
	
	// 1. Baca Template
	src := filepath.Join("templates", string(template))
	texBytes, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("gagal membaca template: %w", err)
	}

	// 2. INJEKSI VARIABLE SISTEM (Otomatisasi)
	// a. Logo
	relLogoPath := filepath.Join("..", "templates", "logo", "logo.jpg")
	data["LOGOPATH"] = filepath.ToSlash(relLogoPath)
	
	// b. Penandatangan (Kades)
	signer := os.Getenv("SIGNER_NAME")
	if signer == "" {
		signer = "........................"
	}
	data["SIGNER.NAME"] = signer

	// c. Waktu (Tanggal Surat & Tahun) - SOLUSI ANTI CRASH
	now := time.Now()
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	monthStr := months[now.Month()]
	
	// Format: "29 Desember 2025" & "2025"
	data["TANGGAL_SURAT"] = fmt.Sprintf("%d %s %d", now.Day(), monthStr, now.Year())
	data["TAHUN"] = fmt.Sprintf("%d", now.Year())

	// 3. LOGIKA KHUSUS PER JENIS SURAT
	// SKTM UMUM: Butuh data Camat, TKSK, dan Hitung Umur
	if template == "sktm_umum.tex" {
		// Injeksi Pejabat
		data["NAMA_CAMAT"] = os.Getenv("NAMA_CAMAT")
		data["NIP_CAMAT"] = os.Getenv("NIP_CAMAT")
		data["NAMA_TKSK"] = os.Getenv("NAMA_TKSK")
		data["NR_TKSK"] = os.Getenv("NR_TKSK")

		// Hitung Umur
		tempat := data["TEMPAT_LAHIR"]
		tglLahirStr := data["TANGGAL_LAHIR"] 
		
		// Parsing Tanggal Lahir (Coba SQL format lalu Indo format)
		var umur int
		parsedDate, err := time.Parse("2006-01-02", tglLahirStr) 
		if err != nil {
			parsedDate, err = time.Parse("02-01-2006", tglLahirStr)
		}

		if err == nil {
			umur = now.Year() - parsedDate.Year()
			if now.YearDay() < parsedDate.YearDay() {
				umur--
			}
			// Re-format tanggal lahir ke Indonesia
			bulanLahir := months[parsedDate.Month()]
			tglLahirIndo := fmt.Sprintf("%d %s %d", parsedDate.Day(), bulanLahir, parsedDate.Year())
			
			// Hasil: "Lampung Timur, 20 Januari 1990 / 34 Tahun"
			data["TTL_UMUR"] = fmt.Sprintf("%s, %s / %d Tahun", tempat, tglLahirIndo, umur)
		} else {
			// Fallback jika gagal parse
			data["TTL_UMUR"] = fmt.Sprintf("%s, %s", tempat, tglLahirStr)
		}
	}

	// 4. Proses Pengisian Template & Penulisan File
	filled := fillTemplate(string(texBytes), data)
	texPath := filepath.Join(tempDir, strings.Replace(pdfName, ".pdf", ".tex", 1))
	pdfPath := filepath.Join(tempDir, pdfName)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder %s: %w", tempDir, err)
	}
	if err := os.WriteFile(texPath, []byte(filled), 0644); err != nil {
		return "", fmt.Errorf("gagal menulis file %s: %w", texPath, err)
	}

	// 5. Eksekusi Background Process (Compile & Kirim)
	go func() {
		absTemp, _ := filepath.Abs(tempDir)
		
		// Jalankan pdflatex
		cmd := exec.Command("pdflatex",
			"-interaction=nonstopmode",
			"-output-directory", absTemp,
			filepath.Base(texPath),
		)
		cmd.Dir = absTemp 

		output, err := cmd.CombinedOutput()

		// Cek keberhasilan (PDF harus ada)
		if _, statErr := os.Stat(pdfPath); os.IsNotExist(statErr) {
			log.Printf("[Surat-Error] pdflatex GAGAL: %v\nOutput:\n%s", err, output)
			_ = SendMessage(client, jid, fmt.Sprintf("❌ Gagal membuat surat %s. Mohon hubungi admin.", NamaSuratmap[template]))
			return
		}
		
		// Log warning jika ada (tapi PDF sukses)
		if err != nil {
			log.Printf("[SURAT-WARN] pdflatex warning: %v", err)
		}
		log.Printf("[SURAT] PDF Berhasil Dibuat: %s", pdfPath)

		// ---------------------------------------------------------
		// TAHAP DISTRIBUSI & LOGGING
		// ---------------------------------------------------------

		// A. Tentukan Nama Pemohon (Handle NAMA vs NAMA_PEMOHON)
		namaPemohon := data["NAMA"]
		if namaPemohon == "" {
			namaPemohon = data["NAMA_PEMOHON"]
		}
		if namaPemohon == "" {
			namaPemohon = "Warga (Tanpa Nama)"
		}

		// B. Kirim ke User (Pemohon)
		caption := fmt.Sprintf("✅ *Surat Selesai Dibuat*\n\n📄 Jenis: %s\n👤 Atas Nama: %s\n🆔 ID: *%s*", NamaSuratmap[template], namaPemohon, uniqueID)
		err = SendFile(client, jid, pdfPath, caption)
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal kirim WA ke user: %v", err)
			_ = SendMessage(client, jid, "File PDF sudah jadi, tapi gagal dikirim via WA. Silakan cek status nanti.")
		}

		// C. Kirim ke Perangkat Desa (Arsip WA)
		nomorKades := os.Getenv("NOMOR_PERANGKAT_DESA")
		if nomorKades != "" {
			kadesJID := fmt.Sprintf("%s@s.whatsapp.net", nomorKades)
			log.Printf("[SURAT] Mengirim salinan ke Perangkat Desa (%s)", kadesJID)
			
			captionKades := fmt.Sprintf("🔔 *Laporan Surat Baru*\n\nJenis: %s\nPemohon: %s\nID: %s\nTanggal: %s", 
				NamaSuratmap[template], namaPemohon, uniqueID, data["TANGGAL_SURAT"])
			
			_ = SendFile(client, kadesJID, pdfPath, captionKades)
		}

		// D. Upload ke Cloudflare R2 (Arsip Cloud)
		log.Println("[SURAT] Mengunggah ke R2...")
		fileURL, err := UploadPDFToR2(pdfPath, pdfName, r2Uploader)
		if err != nil {
			log.Printf("[R2-ERROR] Gagal upload: %v", err)
			fileURL = "Gagal Upload"
		} else {
			log.Printf("[R2] Berhasil upload: %s", fileURL)
		}

		// E. Catat Log ke Google Sheets
		tglLog := time.Now().Format("02-01-2006")
		sheetsClient.AppendSuratLog(string(template), namaPemohon, uniqueID, tglLog, "Belum Diproses", fileURL)

		// F. Bersih-bersih File Sampah
		time.Sleep(2 * time.Second) // Beri jeda agar file tidak terkunci
		os.Remove(texPath)
		os.Remove(pdfPath)
		os.Remove(strings.Replace(texPath, ".tex", ".log", 1))
		os.Remove(strings.Replace(texPath, ".tex", ".aux", 1))
		log.Printf("[SURAT] File sementara %s telah dibersihkan.", pdfName)

	}()

	return texPath, nil
}
