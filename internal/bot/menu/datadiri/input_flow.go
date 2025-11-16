package datadiri

import (
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"log"
)

// bro i change it to switch case, it make nosense to use if else in case like this ~by sina

// handleInputFlow menangani alur 19 langkah
func (h *DataDiriHandler) handleInputFlow(session *db.DataEntrySession, text string) []string {
	stepInfo, ok := Steps[session.CurrentStep]
	if !ok {
		// Jika langkah tidak ada di map (misal > 19), anggap selesai dan pindah ke konfirmasi
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	}

	value, err := ValidateInput(text, stepInfo)
	if err != nil {
		return []string{err.Error()}
	}

	if session.CurrentStep == common.STEP_NIK {
		errMsg, err := h.Service.CheckNIKExists(value.(string), session.JID)
		if err != nil {
			if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_NIK_DUPLIKATE); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{errMsg}
		}
	}

	if err := db.UpdateDataEntrySession(h.Service.DB, session.JID, stepInfo.Field, value); err != nil {
		return []string{"Maaf, terjadi kesalahan saat menyimpan data."}
	}

	if session.CurrentStep == common.STEP_SUKU {
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_CHECKPOINT_SUKU); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Data inti (sampai Suku) sudah tersimpan.\n\n" +
			"Ketik 'cukup' untuk menyelesaikan."}
	}

	nextStep := session.CurrentStep + 1
	if nextStepInfo, ok := Steps[nextStep]; ok {
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, nextStep); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(nextStepInfo)}
	}

	// Selesai (langkah 19 terisi)
	data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
	if err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
}

// handleNikDuplikat menangani alur NIK duplikat
func (h *DataDiriHandler) handleNikDuplikat(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)

	switch normText {
	case "edit nik":
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(Steps[common.STEP_NIK])}

	case "stop":
		if err := db.DeleteDataEntrySession(h.Service.DB, session.JID); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Pendaftaran dibatalkan.\n\n" + common.GetMainMenu()}

	default:
		return []string{"⚠️ Pilihan tidak valid.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
	}
}

// handleCheckpointSuku menangani checkpoint setelah Suku
func (h *DataDiriHandler) handleCheckpointSuku(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)

	switch normText {
	case "cukup":
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}

	case "lanjut":
		nextStep := common.STEP_SUKU + 1

		nextStepInfo, ok := Steps[nextStep]
		if !ok {
			log.Printf("[ERROR] Gagal menemukan step %d (setelah 'lanjut') di map Steps", nextStep)
			data, _ := db.GetFormattedSessionData(h.Service.DB, session.JID)
			db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI)
			return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		}

		if err := db.UpdateStepOnly(h.Service.DB, session.JID, nextStep); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(nextStepInfo)}

	default:
		return []string{
			"Mohon ketik 'lanjut' untuk meneruskan input data (sampai field ke-39) atau 'cukup' untuk menyelesaikan sekarang.",
		}
	}
}

