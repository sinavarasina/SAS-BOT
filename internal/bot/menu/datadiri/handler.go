package datadiri

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
)

// DataDiriHandler mengelola alur data diri
type DataDiriHandler struct {
	Service *Service
}

// NewHandler membuat handler baru untuk data diri
func NewHandler(db *sqlx.DB, sheets *sheets.SheetsClient) *DataDiriHandler {
	service := NewService(db, sheets)
	return &DataDiriHandler{Service: service}
}

// HandleText adalah router utama untuk modul ini
func (h *DataDiriHandler) HandleText(session *db.DataEntrySession, text string) []string {
	switch session.CurrentStep {
	case common.STEP_MENU_DATA_DIRI:
		return h.handleSubmenu(session, text)
	case common.STEP_EDIT_CARI_NIK:
		return h.handleEditCariNik(session, text)
	case common.STEP_NIK_DUPLIKATE:
		return h.handleNikDuplikat(session, text)
	case common.STEP_CHECKPOINT_SUKU:
		return h.handleCheckpointSuku(session, text)
	case common.STEP_KONFIRMASI_DATA_DIRI:
		return h.handleKonfirmasi(session, text)
	case common.STEP_EDIT_DATA_DIRI:
		return h.handleEditFlow(session, text)
	case common.STEP_ULASAN_DATA_DIRI:
		return h.HandleReview(session, text)
	default:
		// Ini menangani langkah 1-19 (input flow)
		return h.handleInputFlow(session, text)
	}
}

// handleSubmenu menangani pilihan "1. Input" atau "2. Edit"
func (h *DataDiriHandler) handleSubmenu(session *db.DataEntrySession, text string) []string {
	switch text {
	case "1": // Input Data Diri
		if err := db.StartNewSession(h.Service.DB, session.JID); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(Steps[common.STEP_DUSUN])}
	case "2": // Edit Data Diri
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_EDIT_CARI_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Silakan masukkan **NIK 16 digit** yang datanya ingin Anda edit:"}
	default:
		return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2, atau ketik 'reset'."}
	}
}

// handleInputFlow menangani alur 19 langkah
func (h *DataDiriHandler) handleInputFlow(session *db.DataEntrySession, text string) []string {
	stepInfo, ok := Steps[session.CurrentStep]
	if !ok {
		return []string{"Terjadi error: Langkah tidak dikenal. Sesi direset."}
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
		return []string{"Data inti (sampai Suku) sudah tersimpan.\n\nKetik 'cukup' untuk menyelesaikan."}
	}

	nextStep := session.CurrentStep + 1
	if nextStepInfo, ok := Steps[nextStep]; ok {
		return []string{FormatQuestion(nextStepInfo)}
	}

	// Selesai (langkah 19 terisi)
	data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
	if err != nil { return []string{"Maaf, terjadi kesalahan sistem."} }
	if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
}

// handleEditCariNik menangani saat user mengirim NIK untuk diedit
func (h *DataDiriHandler) handleEditCariNik(session *db.DataEntrySession, text string) []string {
	data, err := db.GetDataPendudukByNIK(h.Service.DB, text)
	if err != nil {
		log.Printf("[DEBUG] NIK %s tidak ditemukan di DB: %v", text, err)
		return []string{"NIK tidak ditemukan di database. Silakan coba lagi atau ketik 'reset'."}
	}
	if err := db.LoadSessionFromPenduduk(h.Service.DB, session.JID, *data); err != nil {
		log.Printf("[ERROR] Gagal LoadSessionFromPenduduk: %v", err)
		return []string{"NIK ditemukan, tapi gagal memuat data ke sesi. Hubungi admin."}
	}
	dataStr, _ := db.GetFormattedSessionData(h.Service.DB, session.JID)
	// (Set langkah ke Konfirmasi, 42 di skema lama, 20 di skema baru)
	if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{"Data ditemukan. Silakan periksa:\n\n" + dataStr, "\n\nKetik 'valid' untuk menyimpan atau 'edit' untuk mengubah data."}
}

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
		if err != nil { return []string{"maaf, terjadi kesalahan sistem"} }
		if err := db.SetEditField(h.Service.DB, session.JID, ""); err != nil { return []string{"maaf, terjadi kesalahan sistem"} }
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_EDIT_DATA_DIRI); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{"Ketik nomor yang ingin anda edit (1-19)\n\n" + data}
	
	} else {
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil { return []string{"maaf, terjadi kesalahan sistem"} }
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	}
}

// handleEditFlow menangani pengeditan field satu per satu
func (h *DataDiriHandler) handleEditFlow(session *db.DataEntrySession, text string) []string {
	editField := session.EditField.String

	if editField == "" { // User baru masuk mode edit, menunggu nomor
		num, err := strconv.Atoi(text)
		if err != nil || num < 1 || num > 19 {
			data, _ := db.GetFormattedSessionData(h.Service.DB, session.JID)
			return []string{fmt.Sprintf("⚠️ nomor tidak valid\n\nsilakan ketik:\n- nomor 1-19 untuk mengedit data\n- 'valid' untuk menyimpan\n\n%s", data)}
		}
		
		step := Steps[num]
		if err := db.SetEditField(h.Service.DB, session.JID, step.Field); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{fmt.Sprintf("📝 Edit data nomor %d:\n\n%s", num, FormatQuestion(step))}
	
	} else { // User mengirimkan nilai baru untuk field yang dipilih
		step, ok := Steps[session.CurrentStep-1] // (Ini asumsi, lebih baik cari berdasarkan field)
		// Cari step berdasarkan editField
		var currentStep Step
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

		if err := db.SetEditField(h.Service.DB, session.JID, ""); err != nil { return []string{"maaf, terjadi kesalahan sistem"} }
		
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil { return []string{"maaf, terjadi kesalahan sistem"} }
		
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_EDIT_DATA_DIRI); err != nil {
			return []string{"maaf, terjadi kesalahan sistem"}
		}
		return []string{"✅ Data berhasil diupdate\n\nketik:\n- nomor 1-19 untuk mengedit data lain\n- 'valid' jika sudah selesai\n\n" + data}
	}
}

// handleNikDuplikat menangani alur NIK duplikat
func (h *DataDiriHandler) handleNikDuplikat(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)
	if normText == "edit nik" {
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(Steps[common.STEP_NIK])}
	} else if normText == "stop" {
		if err := db.DeleteDataEntrySession(h.Service.DB, session.JID); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Pendaftaran dibatalkan.\n\n" + common.GetMainMenu()}
	} else {
		return []string{"⚠️ Pilihan tidak valid.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
	}
}

// handleCheckpointSuku menangani checkpoint setelah Suku
func (h *DataDiriHandler) handleCheckpointSuku(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)
	if normText == "lanjut" || normText == "cukup" {
		data, err := db.GetFormattedSessionData(h.Service.DB, session.JID)
		if err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	} else {
		return []string{"Mohon ketik 'lanjut' atau 'cukup'."}
	}
}
