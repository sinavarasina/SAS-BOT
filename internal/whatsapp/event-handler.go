// file: internal/whatsapp/event-handler.go
package whatsapp

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
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

			senderJID := v.Info.Sender.String()
			chatJID := v.Info.Chat.String()
			username := v.Info.PushName
			number := senderJID[:strings.Index(senderJID, "@")]


			// 1. Cek apakah ini pesan pribadi (bukan grup)
			if !v.Info.IsGroup {
				// 2. Dapatkan sesi DB pengguna
				session, err := db.GetOrCreateDataEntrySession(appDB, senderJID)
				if err != nil {
					log.Printf("[ERROR] Gagal mendapatkan sesi untuk %s: %v", senderJID, err)
					return
				}

				// 3. Cek apakah pesan ini adalah gambar
				imageMsg := v.Message.GetImageMessage()

				// 4. JIKA INI GAMBAR dan SESI = 300 (Menunggu Pengaduan)
				if imageMsg != nil && session.CurrentStep == bot.STEP_PENGADUAN_WAITING {
					// Panggil handler pengaduan
					bot.HandleImagePengaduan(
						client,
						appDB, // Teruskan appDB
						sheetsClient,
						senderJID,
						imageMsg,
						v.Info.ID,
						v.Info.Chat,
					)
					return // Hentikan proses di sini
				}
			}

			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}

			if v.Info.IsGroup {
				reply := bot.HandlerRouteGroup(appDB, chatJID, text, username, number)
				if reply != "" {
					// (Logika kirim pesan grup...)
					ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					_, err := client.SendMessage(ctxWithTimeout, v.Info.Chat, &waProto.Message{
						Conversation: proto.String(reply),
					})
					cancel()
					if err != nil {
						log.Printf("Error sending group reply to %s: %v", chatJID, err)
					}
				}
			} else {
				// Panggil handler teks pribadi (yang menangani menu 1, 2, 3)
				replies := bot.HandlerRoutePrivate(appDB, chatJID, text, username, number, sheetsClient)
				for _, reply := range replies {
					if reply != "" {
						ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						_, err := client.SendMessage(ctxWithTimeout, v.Info.Chat, &waProto.Message{
							Conversation: proto.String(reply),
						})
						cancel()
						if err != nil {
							log.Printf("Error sending private reply to %s: %v", chatJID, err)
						}
					}
				}
			}
		}
	}
}
