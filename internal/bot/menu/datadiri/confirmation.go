package datadiri

import (
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"log"
)

// handleKonfirmasi menangani 'valid' atau 'edit'
func (h *DataDiriHandler) handleKonfirmasi(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)

	if normText == "valid" {
		if err := h.Service.SaveDataToDBAndSheets(session.JID); err != nil {
			return []string{err.Error()}
		}

		// Pindah ke Alur Ulasan
		if err := db.UpdateSessionField(h.Service.DB, session.JID, "surat_temp_answer", "Input Data Diri"); err != nil {
			log.Printf("[ERROR] Gagal menyimpan nama layanan ulasan: %v", err)
		}
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_ULASAN_DATA_DIRI); err != nil {
			log.Printf("[ERROR] Gagal pindah ke langkah ulasan: %v", err)
		}
		return []string{common.GetUlasanMessage("Input Data Diri")}

	} else if normText == "edit" {
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		if err := db.SetEditField(h.Service.DB, session.JID, ""); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_EDIT_DATA_DIRI); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{"Ketik nomor yang ingin anda edit (1-19)\n\n" + data}

	} else {
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	}
}
