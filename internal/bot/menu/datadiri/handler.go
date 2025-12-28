package datadiri

import (
	"log"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// DataDiriHandler mengelola alur data diri
type DataDiriHandler struct {
	Service *Service
}

// NewHandler membuat handler baru untuk data diri
func NewHandler(ctx *common.ServiceContext) *DataDiriHandler {
	service := NewService(ctx)
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

func (h *DataDiriHandler) HandleImage(session *db.DataEntrySession, imageMsg *waE2E.ImageMessage, messageID string, chatJID types.JID) []string {
	log.Printf("[WARN] Modul DataDiri menerima gambar di langkah %d, ini seharusnya tidak terjadi.", session.CurrentStep)
	return []string{"Maaf, saya tidak mengharapkan gambar saat ini. Mohon kirimkan jawaban dalam bentuk teks."}
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
	// Set langkah ke Konfirmasi
	if err := db.UpdateStepOnly(h.Service.DB, session.JID, common.STEP_KONFIRMASI_DATA_DIRI); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{"Data ditemukan. Silakan periksa:\n\n" + dataStr, "\n\nKetik 'valid' untuk menyimpan atau 'edit' untuk mengubah data."}
}

func (h *DataDiriHandler) handleSubmenu(session *db.DataEntrySession, text string) []string {
	switch text {
	case "1":
		if err := db.StartNewSession(h.Service.DB, session.JID); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{FormatQuestion(Steps[STEP_DUSUN])}
	case "2":
		nik := session.SuratValidNik.String 
		if nik == ""{
			return []string{"Sesi NIK tidak valid silahkan ketik 'reset'."}
		}

		return h.handleEditCariNik(session,nik)
	default:
		return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2, atau ketik 'reset'."}
	}
}
