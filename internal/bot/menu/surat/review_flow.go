package surat

import (
	"log"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

// HandleReview menangani input ulasan
func (h *SuratHandler) HandleReview(session *db.DataEntrySession, text string) []string {
	rating := common.NormalizeInput(text)
	if rating != "1" && rating != "2" && rating != "3" && rating != "4" && rating != "5" {
		return []string{"Input tidak valid. Mohon berikan ulasan berupa angka 1, 2, 3, 4, atau 5."}
	}

	serviceName := session.SuratTempAnswer.String
	if serviceName == "" {
		serviceName = "Pengajuan Surat" // Fallback
	}

	go h.Service.Ctx.SheetsClient.AppendUlasan("ulasan_pengajuan_surat", serviceName, rating, session.JID)

	if err := db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID); err != nil {
		log.Printf("[ERROR] Gagal menghapus sesi setelah ulasan: %v", err)
	}

	return []string{common.GetUlasanThanksMessage()}
}
