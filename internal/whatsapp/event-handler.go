package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func EventHandler(client *whatsmeow.Client, appDB *sqlx.DB, sheetsClient *sheets.SheetsClient) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}

			chatJID := v.Info.Chat.String()
			senderJID := v.Info.Sender.String()
			text := v.Message.GetConversation()
			username := v.Info.PushName
			number := senderJID[:strings.Index(senderJID, "@")]
			
			s := bot.GetSession(senderJID)

			// Cek apakah pesan adalah gambar dan user sedang dalam sesi pengaduan
			imageMsg := v.Message.GetImageMessage()
			if !v.Info.IsGroup && imageMsg != nil && s.Step == "menunggu_pengaduan" {
				
				// 1. download gambar
				data, err := client.Download(context.Background(), imageMsg)
				if err != nil {
					log.Printf("Gagal download gambar dari %s: %v", senderJID, err)
					return
				}

				// 2. path penyimpanan
				os.MkdirAll("uploads", os.ModePerm)
				filePath := fmt.Sprintf("uploads/%s.jpg", v.Info.ID)

				// 3. save pict 
				err = os.WriteFile(filePath, data, 0600)
				if err != nil {
					log.Printf("Gagal menyimpan gambar dari %s: %v", senderJID, err)
					return
				}

				//tambahan kirim ke google drive 
				publicURL, err := uploader.UploadToImgbb(data)
				if err != nil {
					log.Printf("Gagal mengunggah file ke imgbb: %v", err)
					// Beri pesan error ke user
					reply := "Maaf, terjadi kesalahan saat mengunggah gambar. Laporan dibatalkan."
					client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
						Conversation: proto.String(reply),
					})
					bot.ResetSession(senderJID)
					return
				}
				// 4. post ke db 
				aduan := db.Pengaduan{
					UserJID:    senderJID,
					Deskripsi:  imageMsg.GetCaption(), 
					PictPath: publicURL,
				}
				err = db.SavePengaduan(appDB, aduan)
				if err != nil {
					log.Printf("Gagal menyimpan pengaduan ke db: %v", err)
					return
				}

				go sheetsClient.WritePengaduan(aduan)
				//5. konfirmasi pengaduan
				reply := "Terima kasih, pengaduan Anda sudah kami terima dan akan segera diproses."
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					Conversation: proto.String(reply),
				})
				bot.ResetSession(senderJID)
				return
			}

			text = v.Message.GetConversation()
			if v.Info.IsGroup {
				reply := bot.HandlerRouteGroup(appDB, chatJID, text, username, number)
				if reply != "" {
					_, err := client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
						Conversation: proto.String(reply),
					})
					if err != nil {
						log.Printf("Error sending group reply to %s: %v", chatJID, err)
					}
				}
			} else {
				reply := bot.HandlerRoutePrivate(appDB, chatJID, text, username, number)
				if reply != "" {
					_, err := client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
						Conversation: proto.String(reply),
					})
					if err != nil {
						log.Printf("Error sending private reply to %s: %v", chatJID, err)
					}
				}
			}
		}
	}
}
