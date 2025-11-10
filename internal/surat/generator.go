
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
)

func fillTemplate(content string, data map[string]string) string {
	for key, value := range data {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
}

func GenerateAsync(template JenisSurat, data map[string]string, tempDir string, jid string, client *whatsmeow.Client) (string, error) {
	src := filepath.Join("templates", string(template))
	texBytes, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("gagal membaca template: %w", err)
	}

	// --- PERBAIKAN DI SINI ---
	// Gunakan path RELATIF dari folder 'temp' ke folder 'templates'.
	// Kita gunakan filepath.Join agar aman di semua OS, lalu ToSlash agar LaTeX (/) senang.
	relLogoPath := filepath.Join("..", "templates", "logo", "logo.jpg")
	data["LOGOPATH"] = filepath.ToSlash(relLogoPath)
	// --- AKHIR PERBAIKAN ---

	filled := fillTemplate(string(texBytes), data)
	fileName := fmt.Sprintf("%s_%d.tex", strings.TrimSuffix(string(template), ".tex"), time.Now().Unix())
	texPath := filepath.Join(tempDir, fileName)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder %s: %w", tempDir, err)
	}
	if err := os.WriteFile(texPath, []byte(filled), 0644); err != nil {
		return "", fmt.Errorf("gagal menulis file %s: %w", texPath, err)
	}

	go func() {
		absTemp, _ := filepath.Abs(tempDir)
		pdfPath := strings.Replace(texPath, ".tex", ".pdf", 1)
		
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

		err = SendFile(client, jid, pdfPath,
			fmt.Sprintf("Surat %s telah selesai dibuat dan siap dicetak.", template))
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal mengirim file PDF: %v", err)
			_ = SendMessage(client, jid,
				"File berhasil dibuat tapi gagal dikirim. Silakan hubungi admin.")
		}
		
		// (Opsional tapi disarankan) Hapus file sementara setelah dikirim
		// os.Remove(texPath)
		// os.Remove(pdfPath)
		// os.Remove(strings.Replace(texPath, ".tex", ".log", 1))
		// os.Remove(strings.Replace(texPath, ".tex", ".aux", 1))

	}()

	return texPath, nil
}
