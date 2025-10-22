package bot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func HandleImagePengaduan(
	client *whatsmeow.Client,
	appDB *sqlx.DB,
	sheetsClient *sheets.SheetsClient,
	senderJID string,
	imageMsg *waProto.ImageMessage,
	messageID string,
	chatJID types.JID,
) {
	// 1. download gambar
	data, err := client.Download(context.Background(), imageMsg)
	if err != nil {
		log.Printf("Gagal download gambar dari %s: %v", senderJID, err)
		return
	}

	// 2. simpan ke lokal
	os.MkdirAll("uploads", os.ModePerm)
	filePath := fmt.Sprintf("uploads/%s.jpg", messageID)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		log.Printf("Gagal menyimpan gambar dari %s: %v", senderJID, err)
		return
	}

	// 3. upload ke imgbb
	publicURL, err := uploader.UploadToImgbb(data)
	if err != nil {
		log.Printf("Gagal mengunggah file ke imgbb: %v", err)
		reply := "Maaf, terjadi kesalahan saat mengunggah gambar. Laporan dibatalkan."
		client.SendMessage(context.Background(), chatJID, &waProto.Message{
			Conversation: proto.String(reply),
		})
		ResetSession(senderJID)
		return
	}

	// 4. simpan ke database
	aduan := db.Pengaduan{
		UserJID:   senderJID,
		Deskripsi: imageMsg.GetCaption(),
		PictPath:  publicURL,
	}
	if err := db.SavePengaduan(appDB, aduan); err != nil {
		log.Printf("Gagal menyimpan pengaduan ke db: %v", err)
		return
	}

	// 5. sinkronisasi ke Google Sheets
	go sheetsClient.WritePengaduan(aduan)

	// 6. balas ke user
	reply := "Terima kasih, pengaduan Anda sudah kami terima dan akan segera diproses."
	client.SendMessage(context.Background(), chatJID, &waProto.Message{
		Conversation: proto.String(reply),
	})

	ResetSession(senderJID)
}

