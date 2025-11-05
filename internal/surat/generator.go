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
		cmd.Dir = absTemp

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[SURAT-ERROR] pdflatex gagal: %v\n%s", err, output)
			_ = SendMessage(client, jid,
				fmt.Sprintf("Surat *%s* gagal dikompilasi.", template))
			return
		}

		log.Printf("[SURAT] PDF selesai dibuat: %s", pdfPath)

		err = SendFile(client, jid, pdfPath,
			fmt.Sprintf("Surat *%s* telah selesai dibuat dan siap dicetak.", template))
		if err != nil {
			log.Printf("[SURAT-ERROR] Gagal mengirim file PDF: %v", err)
			_ = SendMessage(client, jid,
				"File berhasil dibuat tapi gagal dikirim. Silakan hubungi admin.")
		}
	}()

	return texPath, nil
}
