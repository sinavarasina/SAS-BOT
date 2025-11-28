package datadiri

import (
	"log"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

// HandleReview menangani input ulasan
func (h *DataDiriHandler) HandleReview(session *db.DataEntrySession, text string) []string {
	rating := common.NormalizeInput(text)
	if rating != "1" && rating != "2" && rating != "3" && rating != "4" && rating != "5" {
		return []string{"Input tidak valid. Mohon berikan ulasan berupa angka 1, 2, 3, 4, atau 5."}
	}

	// Ambil nama layanan yang disimpan
	serviceName := session.SuratTempAnswer.String
	if serviceName == "" {
		serviceName = "Input Data Diri" // Fallback
	}

	// Kirim ulasan ke Google Sheet di background
	go h.Service.SheetsClient.AppendUlasan("ulasan_input_diri", serviceName, rating, session.JID)

	if err := db.ResetSessionToMainMenu(h.Service.DB, session.JID); err != nil {
    log.Printf("[ERROR] Gagal reset sesi: %v", err)
	}
// Pesan konfirmasi + Menu Utama
	return []string{
    "Terima kasih! Ulasan Anda sangat berarti. ⭐",
    common.GetMainMenu(), // Tampilkan menu lagi
	}
}
