package pengaduan

import (
	"log"
	"strings"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type PengaduanHandler struct {
	Service *Service
}

func NewHandler(ctx *common.ServiceContext) *PengaduanHandler {
	service := NewService(ctx)
	return &PengaduanHandler{Service: service}
}

func (h *PengaduanHandler) HandleText(session *db.DataEntrySession, text string) []string {
	switch session.CurrentStep {
	case common.STEP_PENGADUAN_MENU:
		return h.handleMenuUtama(session, text)
	case common.STEP_PENGADUAN_VALIDASI_NIK:
		return h.handleValidasiNik(session, text)
	case common.STEP_PENGADUAN_CARI_ID:
		return h.handleCariID(session, text)
	case common.STEP_ULASAN_PENGADUAN:
		return h.HandleReview(session, text)
	case common.STEP_PENGADUAN_WAITING:
		return []string{"Mohon kirimkan *gambar* pengaduan, bukan teks. Atau ketik 'reset' untuk batal."}
	default:
		log.Printf("[WARN] Alur Pengaduan menerima langkah tidak dikenal: %d", session.CurrentStep)
		db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID)
		return []string{"Terjadi error pada alur pengaduan. Sesi direset."}
	}
}

func (h *PengaduanHandler) HandleImage(session *db.DataEntrySession, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID) []string {
	if session.CurrentStep == common.STEP_PENGADUAN_WAITING {
		return h.Service.HandleImagePengaduan(session, imageMsg, messageID, chatJID)
	}
	log.Printf("[WARN] Modul Pengaduan menerima gambar di langkah %d (seharusnya tidak).", session.CurrentStep)
	return []string{"Maaf, saya tidak mengharapkan gambar saat ini."}
}

func (h *PengaduanHandler) handleMenuUtama(session *db.DataEntrySession, text string) []string {
	switch text {
	case "1":
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_PENGADUAN_VALIDASI_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Untuk mengajukan pengaduan, silakan masukkan **NIK 16 digit** Anda untuk verifikasi:"}
	case "2":
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_PENGADUAN_CARI_ID); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Silakan masukkan *ID Pengaduan* Anda (contoh: P-101):"}
	default:
		return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2."}
	}
}

func (h *PengaduanHandler) handleValidasiNik(session *db.DataEntrySession, text string) []string {
	if len(text) != 16 {
		return []string{"Format NIK salah. Harap masukkan 16 digit NIK Anda:"}
	}
	isRegistered, err := db.CheckNIKExistsInPenduduk(h.Service.Ctx.DB, text)
	if err != nil {
		log.Printf("[ERROR] Gagal mengecek NIK di DB permanen: %v", err)
		return []string{"Maaf, terjadi kesalahan sistem saat validasi NIK."}
	}
	if isRegistered {
		if err := db.UpdateStepOnly(h.Service.Ctx.DB, session.JID, common.STEP_PENGADUAN_WAITING); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"NIK Anda terverifikasi. Silakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."}
	} else {
		if err := db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID); err != nil {
			log.Printf("[ERROR] Gagal hapus sesi setelah NIK tidak ditemukan: %v", err)
		}
		return []string{"NIK Anda tidak terdaftar di sistem kami.\n\nSilakan pilih *Menu 1 (Data Diri)* untuk melakukan input data diri terlebih dahulu sebelum membuat pengaduan.\n\n" + common.GetMainMenu()}
	}
}

func (h *PengaduanHandler) handleCariID(session *db.DataEntrySession, text string) []string {
	unikID := strings.ToUpper(text)
	statusMsg, err := h.Service.GetPengaduanStatus(unikID)
	if err != nil {
		return []string{err.Error()}
	}
	if err := db.DeleteDataEntrySession(h.Service.Ctx.DB, session.JID); err != nil {
		log.Printf("[ERROR] Gagal hapus sesi setelah cek status: %v", err)
	}
	return []string{statusMsg}
}
