package surat

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// SuratHandler mengelola alur pengajuan surat
type SuratHandler struct {
	Service *Service
}

// NewHandler membuat handler baru untuk surat
func NewHandler(ctx *common.ServiceContext) *SuratHandler {
	service := NewService(ctx)
	return &SuratHandler{Service: service}
}

// HandleText adalah router utama untuk modul ini
func (h *SuratHandler) HandleText(session *db.DataEntrySession, text string) []string {
	switch session.CurrentStep {
	case common.STEP_SURAT_MENU_UTAMA:
		return h.handleMenuUtama(session, text)
	case common.STEP_SURAT_VALIDASI_NIK:
		return h.handleValidasiNik(session, text)
	case common.STEP_SURAT_PILIH_JENIS:
		return h.handlePilihJenis(session, text)
	case common.STEP_SURAT_INPUT_DATA:
		return h.handleInputFlow(session, text)
	case common.STEP_SURAT_KONFIRMASI_FIELD:
		return h.handleKonfirmasiField(session, text)
	case common.STEP_SURAT_CEK_PROGRES:
		return h.handleCekProgres(session, text)
	case common.STEP_ULASAN_SURAT:
		return h.HandleReview(session, text)
	default:
		log.Printf("[WARN] Alur Surat menerima langkah tidak dikenal: %d", session.CurrentStep)
		db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID)
		return []string{"Terjadi error pada alur surat. Sesi direset."}
	}
}

// HandleImage (tidak digunakan untuk modul ini)
func (h *SuratHandler) HandleImage(session *db.DataEntrySession, imageMsg *waE2E.ImageMessage, messageID string, chatJID types.JID) []string {
	log.Printf("[WARN] Modul Surat menerima gambar, ini seharusnya tidak terjadi.")
	return []string{"Maaf, saya tidak mengharapkan gambar saat ini. Mohon kirimkan jawaban dalam bentuk teks."}
}

// handleMenuUtama menangani "1. Ajukan" atau "2. Cek"
func (h *SuratHandler) handleMenuUtama(session *db.DataEntrySession, text string) []string {
	switch text {
	case "1": // 1. Ajukan Surat
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_VALIDASI_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Baik, silakan masukkan **NIK 16 digit** Anda untuk validasi data:"}
	case "2": // 2. Cek Progres Surat
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_CEK_PROGRES); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Silakan masukkan *Nomor Unik Surat* Anda (contoh: 1234):"}
	default:
		return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2."}
	}
}

// handleValidasiNik memvalidasi NIK dan meng-auto-fill data
func (h *SuratHandler) handleValidasiNik(session *db.DataEntrySession, text string) []string {
	dataPenduduk, err := db.GetDataPendudukByNIK(h.Service.Ctx.DB, text)
	if err != nil {
		if err := db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID); err != nil { /*...*/
		}
		return []string{"NIK Anda tidak terdaftar. Silakan pilih Menu 1 untuk Input Data Diri dahulu.\n\n" + common.GetMainMenu()}
	}

	if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_valid_nik", text); err != nil {
		return []string{"Kesalahan menyimpan NIK. Coba lagi."}
	}

	dataMap := BuildDataMap(db.DataPenduduk(*dataPenduduk))
	dataMapBytes, _ := json.Marshal(dataMap)
	if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_data_map", string(dataMapBytes)); err != nil {
		return []string{"Kesalahan menyimpan data map. Coba lagi."}
	}

	if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_PILIH_JENIS); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{
		"NIK Tervalidasi. Silakan pilih jenis surat:\n" +
			"1. Surat Domisili\n2. Surat Usaha\n3. SKTM Umum\n4. SKTM Tanggungan\n5. Surat Kematian",
	}
}

// handlePilihJenis menangani pemilihan 1-5
func (h *SuratHandler) handlePilihJenis(session *db.DataEntrySession, text string) []string {
	jenisSurat, ok := JenisSuratMap[text]
	if !ok {
		return []string{"Pilihan tidak valid. Masukkan angka 1-5."}
	}
	if err := db.SetEditField(h.Service.Ctx.DB, session.JID, string(jenisSurat)); err != nil { /*...*/
	}

	nik, _ := db.GetSessionField(h.Service.Ctx.DB, session.JID, "surat_valid_nik")
	dataPenduduk, _ := db.GetDataPendudukByNIK(h.Service.Ctx.DB, nik.String)

	fieldList := GetFieldList(db.DataPenduduk(*dataPenduduk), jenisSurat)

	if len(fieldList) == 0 {
		return h.Service.HandleSuratGeneration(session)
	}

	fieldStr := strings.Join(fieldList, ",")
	if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_fields_pending", fieldStr); err != nil { /*...*/
	}
	if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_field_now", fieldList[0]); err != nil { /*...*/
	}
	if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_INPUT_DATA); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}

	return []string{GetPrompt(fieldList[0])}
}

// handleCekProgres menangani permintaan status
func (h *SuratHandler) handleCekProgres(session *db.DataEntrySession, text string) []string {
	unikID := strings.ToUpper(text)
	statusMsg, err := h.Service.GetSuratStatus(unikID)
	if err != nil {
		return []string{err.Error()}
	}
	if err := db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID); err != nil { /*...*/
	}
	return []string{statusMsg}
}
