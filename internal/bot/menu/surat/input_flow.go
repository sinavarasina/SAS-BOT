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
	// 2. Simpan jawaban sementara
	if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_temp_answer", text); err != nil {
		return []string{"Kesalahan menyimpan jawaban sementara."}
	}
	
	// 3. Lanjut ke konfirmasi
	if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_KONFIRMASI_FIELD); err != nil {
		return []string{"Maaf, terjadi kesalahan sistem."}
	}
	return []string{fmt.Sprintf("Anda mengisi: *%s*\n\nKetik 'ya' untuk lanjut, atau 'edit' untuk mengulangi.", text)}
}

// handleKonfirmasiField menangani "ya" atau "edit"
func (h *SuratHandler) handleKonfirmasiField(session *db.DataEntrySession, text string) []string {
	normText := common.NormalizeInput(text)
	currentField := session.SuratFieldNow.String
	fieldList := strings.Split(session.SuratFieldsPending.String, ",")

	if normText == "edit" {
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_INPUT_DATA); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{GetPrompt(currentField)} // Tanyakan lagi
	}

	if normText == "ya" {
		// Jawaban "ya", simpan data
		tempAnswer := session.SuratTempAnswer.String
		dataMap := make(map[string]string)
		_ = json.Unmarshal([]byte(session.SuratDataMap.String), &dataMap)
		dataMap[currentField] = tempAnswer

		next := NextField(fieldList, currentField)

		if next != "" { // Masih ada pertanyaan
			dataMapBytes, _ := json.Marshal(dataMap)
			if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_data_map", string(dataMapBytes)); err != nil {
				return []string{"Kesalahan menyimpan data map."}
			}
			if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_field_now", next); err != nil {
				return []string{"Kesalahan menyimpan field berikutnya."}
			}
			if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_SURAT_INPUT_DATA); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{GetPrompt(next)}
		}

		// Selesai, buat surat
		dataMapBytes, _ := json.Marshal(dataMap)
		if err := db.UpdateSessionField(h.Service.Ctx.DB, session.JID, "surat_data_map", string(dataMapBytes)); err != nil {
			return []string{"Kesalahan menyimpan data map akhir."}
		}
		return h.Service.HandleSuratGeneration(session)
	}

	return []string{"Pilihan tidak valid. Ketik 'ya' atau 'edit'."}
}
