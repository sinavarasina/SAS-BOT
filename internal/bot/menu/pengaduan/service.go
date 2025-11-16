// file: internal/bot/menu/pengaduan/service.go
package pengaduan

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// "github.com/jmoiron/sqlx" // <-- (Diperlukan oleh db)
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	// "github.com/sinavarasina/SAS-BOT/internal/uploader" // <-- (Diperlukan oleh s.Ctx.ImgbbUploader)
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto" // <-- Diberi nama 'waProto'
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto" // <-- PERBAIKAN: Import 'proto'
)

// Service menangani logika bisnis untuk modul Pengaduan
type Service struct {
	Ctx *common.ServiceContext
}

// NewService membuat instance service baru
func NewService(ctx *common.ServiceContext) *Service {
	return &Service{Ctx: ctx}
}

// HandleImagePengaduan adalah fungsi utama yang memproses gambar pengaduan
func (s *Service) HandleImagePengaduan(session *db.DataEntrySession, imageMsg *waProto.ImageMessage, messageID string, chatJID types.JID) []string {
	log.Printf("[DEBUG] Processing image complaint from %s", session.JID)

	data, err := s.Ctx.WAClient.Download(context.Background(), imageMsg)
	if err != nil {
		log.Printf("[ERROR] Failed to download image from %s: %v", session.JID, err)
		return []string{"Gagal mengunduh gambar. Mohon coba lagi."}
	}

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Printf("[ERROR] Failed to create uploads directory: %v", err)
		return []string{"Terjadi kesalahan sistem (gagal buat folder)."}
	}
	filePath := fmt.Sprintf("uploads/%s.jpg", messageID)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		log.Printf("[ERROR] Failed to save image from %s: %v", session.JID, err)
		return []string{"Terjadi kesalahan saat menyimpan gambar."}
	}

	go func() {
		start := time.Now()
		log.Printf("[ASYNC] Starting upload & synchronization for %s", session.JID)

		// --- PERBAIKAN: Panggil method Uploader ---
		publicURL, err := s.Ctx.ImgbbUploader.UploadToImgbb(data)
		if err != nil {
			log.Printf("[ERROR] Failed to upload to ImgBB: %v", err)
			sendMessageSafe(s.Ctx.WAClient, chatJID, "Maaf, terjadi kesalahan saat mengunggah gambar. Silakan coba lagi.")
			db.DeleteDataEntrySession(s.Ctx.DB, session.JID)
			return
		}
		
		os.Remove(filePath)

		aduan := db.Pengaduan{
			UserJID:   session.JID,
			Deskripsi: imageMsg.GetCaption(),
			PictPath:  publicURL,
		}
		newID, err := db.SavePengaduan(s.Ctx.DB, aduan)
		if err != nil {
			log.Printf("[ERROR] Failed to save complaint to database: %v", err)
			sendMessageSafe(s.Ctx.WAClient, chatJID, "Terjadi kesalahan saat menyimpan laporan Anda.")
			return
		}

		publicID := fmt.Sprintf("P-%d", 100+newID)

		go s.Ctx.SheetsClient.AppendPengaduan(aduan, publicID)

		if err := db.UpdateSessionField(s.Ctx.DB, session.JID, "surat_temp_answer", "Ajukan Pengaduan"); err != nil {
			log.Printf("[ERROR] Gagal menyimpan nama layanan ulasan: %v", err)
		}
		if err := db.UpdateStepOnly(s.Ctx.DB, session.JID, common.STEP_ULASAN_PENGADUAN); err != nil {
			log.Printf("[ERROR] Gagal pindah ke langkah ulasan: %v", err)
		}
		
		ulasanMsg := fmt.Sprintf("Pengaduan Anda telah tersimpan.\nNomor ID Pengaduan Anda adalah: *%s*", publicID) +
			"\n\n" + common.GetUlasanMessage("Ajukan Pengaduan")
		sendMessageSafe(s.Ctx.WAClient, chatJID, ulasanMsg)

		log.Printf("[DONE] Complaint from %s processed in %v (URL: %s)", session.JID, time.Since(start), publicURL)
	}()

	return []string{"Terima kasih, pengaduan Anda sudah kami terima dan sedang diproses..."}
}

// GetPengaduanStatus mengambil status dari sheets
func (s *Service) GetPengaduanStatus(unikID string) (string, error) {
	status, err := s.Ctx.SheetsClient.GetPengaduanStatus(unikID)
	if err != nil {
		log.Printf("[WARN] Gagal mencari status untuk ID %s: %v", unikID, err)
		return "", fmt.Errorf("ID Pengaduan *%s* tidak ditemukan.", unikID)
	}
	return fmt.Sprintf("Status untuk pengaduan *%s*:\n\n*STATUS: %s*", unikID, status), nil
}

// (Helper kirim pesan)
func sendMessageSafe(client *whatsmeow.Client, chat types.JID, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// --- PERBAIKAN: Gunakan 'waProto.Message' ---
	_, err := client.SendMessage(ctx, chat, &waProto.Message{
		Conversation: proto.String(text), // <-- 'proto' sekarang sudah dikenal
	})
	if err != nil {
		log.Printf("[WARN] Failed to send message to %s: %v", chat.String(), err)
	}
}
