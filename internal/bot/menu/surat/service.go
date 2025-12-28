package surat

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

// Service menangani logika bisnis untuk modul Surat
type Service struct {
	Ctx *common.ServiceContext
}

// NewService membuat instance service baru
func NewService(ctx *common.ServiceContext) *Service {
	return &Service{Ctx: ctx}
}

// HandleSuratGeneration adalah fungsi utama yang memproses pembuatan surat
func (s *Service) HandleSuratGeneration(session *db.DataEntrySession) []string {
	// 1. Ambil data sesi terbaru (termasuk data map yang baru diupdate)
	fullSession, err := db.GetFullSessionData(s.Ctx.DB, session.JID)
	if err != nil {
		log.Printf("[SURAT-ERROR] Gagal mengambil data sesi lengkap: %v", err)
		return []string{"Terjadi kesalahan saat mengambil data sesi Anda."}
	}

	// 2. Unmarshal data map yang sudah terkumpul
	dataMap := make(map[string]string)
	if err := json.Unmarshal([]byte(fullSession.SuratDataMap.String), &dataMap); err != nil {
		log.Printf("[SURAT-ERROR] Gagal unmarshal data map: %v", err)
		return []string{"Terjadi kesalahan data map. Silakan 'reset'."}
	}

	jenisStr := fullSession.EditField.String // (Misal: "sk_domisili.tex")
	jenis := JenisSurat(jenisStr)
	namaSurat := NamaSuratmap[jenis]

	// 3. Buat ID unik & nama file
	unikID := fmt.Sprintf("%04d", 1000+rand.Intn(9000)) // 4 digit random
	tgl := time.Now().Format("02-01-2006")
	pdfName := fmt.Sprintf("%s_%s_%s.pdf", strings.TrimSuffix(string(jenis), ".tex"), unikID, tgl)

	// 4. Panggil generator (akan berjalan di background)
	_, err = GenerateAsync(
		jenis, dataMap, "temp", session.JID,
		s.Ctx.WAClient, pdfName, unikID,
		s.Ctx.SheetsClient, s.Ctx.R2Uploader,
	)
	if err != nil {
		log.Printf("[SURAT-ERROR] %v", err)
		return []string{"Terjadi kesalahan saat memproses surat."}
	}

	// 5. Pindah ke alur ulasan
	if err := db.UpdateSessionField(s.Ctx.DB, session.JID, "surat_temp_answer", namaSurat); err != nil {
		log.Printf("[ERROR] Gagal menyimpan nama layanan ulasan: %v", err)
	}
	if err := db.UpdateStepOnly(s.Ctx.DB, session.JID, common.STEP_ULASAN_SURAT); err != nil {
		log.Printf("[ERROR] Gagal pindah ke langkah ulasan: %v", err)
	}

	infoMsg := fmt.Sprintf(
		"✅ *Permintaan Diterima*\n\n"+
			"📄 Jenis: %s\n"+
			"🆔 *ID Surat: %s*\n\n"+
			"Mohon tunggu sebentar, sistem sedang membuat dan mengirimkan file PDF surat Anda...",
		namaSurat, unikID,
	)

	return []string{infoMsg, common.GetUlasanMessage(namaSurat)}
}

// GetSuratStatus mengambil status dari sheets
func (s *Service) GetSuratStatus(unikID string) (string, error) {
	status, err := s.Ctx.SheetsClient.GetSuratStatus(unikID)
	if err != nil {
		log.Printf("[WARN] Gagal mencari status untuk ID %s: %v", unikID, err)
		return "", fmt.Errorf("ID Surat *%s* tidak ditemukan.", unikID)
	}
	return fmt.Sprintf("Status untuk surat *%s*:\n\n*STATUS: %s*", unikID, status), nil
}
