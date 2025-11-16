package datadiri

import (
	"fmt"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"strconv"
)

// handleEditFlow menangani pengeditan field satu per satu
func (h *DataDiriHandler) handleEditFlow(session *db.DataEntrySession, text string) []string {
	editField := session.EditField.String

	// Skenario 1: User baru masuk mode edit, menunggu nomor
	if editField == "" {
		num, err := strconv.Atoi(text)
		if err != nil || num < 1 || num > 39 {
			data, _ := db.GetFormattedSessionData(h.Service.DB, session.JID)
			return []string{fmt.Sprintf("⚠️ nomor tidak valid\n\nsilakan ketik:\n- nomor 1-39 untuk mengedit data\n- 'valid' untuk menyimpan\n\n%s", data)} // <-- PERBAIKAN: 41 -> 19
		}

		step := Steps[num]
		if err := db.SetEditField(h.Service.DB, session.JID, step.Field); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{fmt.Sprintf("📝 Edit data nomor %d:\n\n%s", num, FormatQuestion(step))}
	}

	// Skenario 2: User mengirimkan nilai baru untuk field yang dipilih
	var currentStep Step
	for i, s := range Steps {
		if s.Field == editField {
			currentStep = Steps[i+1] // (Ini mungkin salah, kita cari berdasarkan field)
			break
		}
	}
	// Cari step yang benar berdasarkan field
	for _, s := range Steps {
		if s.Field == editField {
			currentStep = s
			break
		}
	}


	value, err := ValidateInput(text, currentStep)
	if err != nil {
		return []string{fmt.Sprintf("📝 Mode Edit Aktif:\n\n%s", err.Error())}
	}

	// Update nilai di DB
	query := fmt.Sprintf("UPDATE data_entry_sessions SET %s = $1, updated_at = NOW() WHERE jid = $2", editField)
	if _, err := h.Service.DB.Exec(query, value, session.JID); err != nil {
		return []string{"maaf, terjadi kesalahan sistem saat update."}
	}

	if err := db.SetEditField(h.Service.DB, session.JID, ""); err != nil {
		return []string{"maaf, terjadi kesalahan sistem"}
	}
	data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
	if err != nil {
		return []string{"maaf, terjadi kesalahan sistem"}
	}

	if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_EDIT_DATA_DIRI); err != nil {
		return []string{"maaf, terjadi kesalahan sistem"}
	}
	return []string{
		"✅ Data berhasil diupdate\n\nketik:\n- nomor 1-39 untuk mengedit data lain\n- 'valid' jika sudah selesai\n\n" + data, // <-- PERBAIKAN: 41 -> 19
	}
}
