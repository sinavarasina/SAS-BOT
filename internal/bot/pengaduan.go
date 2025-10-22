package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// i make it async but not fully async, atleast some of blocking code i run it at goroutine
// need fully async implementation
func HandleImagePengaduan(
	client *whatsmeow.Client,
	appDB *sqlx.DB,
	sheetsClient *sheets.SheetsClient,
	senderJID string,
	imageMsg *waProto.ImageMessage,
	messageID string,
	chatJID types.JID,
) {
	log.Printf("[DEBUG] Processing image complaint from %s", senderJID)

	// Download image from WhatsApp
	data, err := client.Download(context.Background(), imageMsg)
	if err != nil {
		log.Printf("[ERROR] Failed to download image from %s: %v", senderJID, err)
		sendMessageSafe(client, chatJID, "Gagal mengunduh gambar. Mohon coba lagi.")
		return
	}

	// Save temporarily to local storage
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Printf("[ERROR] Failed to create uploads directory: %v", err)
		sendMessageSafe(client, chatJID, "Terjadi kesalahan sistem saat memproses laporan Anda.")
		return
	}

	filePath := fmt.Sprintf("uploads/%s.jpg", messageID)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		log.Printf("[ERROR] Failed to save image from %s: %v", senderJID, err)
		sendMessageSafe(client, chatJID, "Terjadi kesalahan saat menyimpan gambar.")
		return
	}

	// Send instant acknowledgment to user
	sendMessageSafe(client, chatJID, "Terima kasih, pengaduan Anda sudah kami terima dan sedang diproses...")

	// Run heavy tasks asynchronously
	go func() {
		start := time.Now()
		log.Printf("[ASYNC] Starting upload & synchronization for %s", senderJID)

		// Upload to ImgBB
		publicURL, err := uploader.UploadToImgbb(data)
		if err != nil {
			log.Printf("[ERROR] Failed to upload to ImgBB: %v", err)
			sendMessageSafe(client, chatJID, "Maaf, terjadi kesalahan saat mengunggah gambar. Silakan coba lagi.")
			ResetSession(senderJID)
			return
		}

		// Save to local database
		aduan := db.Pengaduan{
			UserJID:   senderJID,
			Deskripsi: imageMsg.GetCaption(),
			PictPath:  publicURL,
		}
		if err := db.SavePengaduan(appDB, aduan); err != nil {
			log.Printf("[ERROR] Failed to save complaint to database: %v", err)
			sendMessageSafe(client, chatJID, "Terjadi kesalahan saat menyimpan laporan Anda.")
			return
		}

		// Write to Google Sheets asynchronously
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WARN] Panic recovered in sheets writer: %v", r)
				}
			}()
			log.Printf("[ASYNC] Writing complaint to Google Sheets for %s", senderJID)
			sheetsClient.WritePengaduan(aduan)
			log.Printf("[DONE] Successfully written complaint to Google Sheets for %s", senderJID)
		}()

		// Notify user after processing is completed
		sendMessageSafe(client, chatJID, "Pengaduan Anda telah tersimpan dan akan segera ditindaklanjuti. Terima kasih atas laporannya!")

		log.Printf("[DONE] Complaint from %s processed in %v (URL: %s)", senderJID, time.Since(start), publicURL)
		ResetSession(senderJID)
	}()
}

// sendMessageSafe sends a message with a safe timeout to prevent hanging
func sendMessageSafe(client *whatsmeow.Client, chat types.JID, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.SendMessage(ctx, chat, &waProto.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		log.Printf("[WARN] Failed to send message to %s: %v", chat.String(), err)
	}
}
