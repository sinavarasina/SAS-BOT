package surat

import (
	"encoding/json"
	"fmt"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"strings"
)

// handleInputFlow menangani pengiriman pertanyaan
func (h *SuratHandler) handleInputFlow(session *db.DataEntrySession, text string) []string {
	// 1. Ambil field yang sedang diisi saat ini
	currentField, _ := db.GetSessionField(h.Service.Ctx.DB, session.JID, "surat_field_now")
	fieldName := currentField.String

	// 2. Ambil Data Map yang sudah ada, lalu update dengan input user
	dataMapJSON, _ := db.GetSessionField(h.Service.Ctx.DB, session.JID, "surat_data_map")
	var dataMap map[string]string
	json.Unmarshal([]byte(dataMapJSON.String), &dataMap)

	// Simpan input user ke map
	dataMap[fieldName] = text

	// Simpan balik map ke database
	updatedMap, _ := json.Marshal(dataMap)
	db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_data_map", string(updatedMap))

	// 3. Cek apakah masih ada field selanjutnya?
	pendingFieldsStr, _ := db.GetSessionField(h.Service.Ctx.DB, session.JID, "surat_fields_pending")
	fields := strings.Split(pendingFieldsStr.String, ",")

	nextFieldName := NextField(fields, fieldName)

	if nextFieldName != "" {
		// KASUS A: Masih ada field berikutnya
		// Update pointer field sekarang ke field berikutnya
		db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_field_now", nextFieldName)
		
		// Langsung tanya pertanyaan berikutnya
		return []string{GetPrompt(nextFieldName)}
	}

	// KASUS B: Semua field sudah terisi -> Tampilkan REKAP FINAL
	// Pindah ke step konfirmasi akhir
	db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_KONFIRMASI_FIELD)

	// Buat ringkasan data agar user bisa cek sebelum cetak
	summary := "*KONFIRMASI DATA SURAT*\nMohon periksa data berikut:\n\n"
	
	// Kita iterate sesuai urutan field yang diminta (fields) agar rapi
	for _, k := range fields {
		val := dataMap[k]
		// Bersihkan underscore biar enak dibaca (opsional)
		label := strings.ReplaceAll(k, "_", " ")
		summary += fmt.Sprintf("• *%s*: %s\n", label, val)
	}

	summary += "\nKetik *'Lanjut'* untuk memproses surat, atau *'Batal'* untuk mengulangi dari awal."
	
	return []string{summary}
}

func (h *SuratHandler) handleKonfirmasiFinal(session *db.DataEntrySession, text string) []string {
	input := strings.ToLower(strings.TrimSpace(text))

	if input == "lanjut" || input == "ya" || input == "ok" {
		// User setuju, GENERATE SURAT SEKARANG
		return h.Service.HandleSuratGeneration(session)
	} else if input == "batal" || input == "reset" {
		// User ingin ulang
		db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID)
		return []string{"Pengajuan dibatalkan. Silakan ketik menu untuk mulai lagi."}
	} else {
		return []string{"Jawaban tidak dikenali. Ketik *'Lanjut'* untuk proses atau *'Batal'* untuk membatalkan."}
	}
}
