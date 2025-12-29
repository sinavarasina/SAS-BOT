package whatsapp

import (
	"context"
	"log"
	"strings"
	// "time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/bot/router"
	"github.com/sinavarasina/SAS-BOT/internal/utils"
	"go.mau.fi/whatsmeow"
	// waProto "go.mau.fi/whatsmeow/binary/proto"
	// "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	// googleProto "google.golang.org/protobuf/proto"
)

func EventHandler(botRouter *router.BotRouter, appDB *sqlx.DB, limiter *utils.RateLimiter) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		
		// 1. Menangani Pesan Normal
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			
			senderJID := v.Info.Sender.String()

			if !limiter.Allow(senderJID) {
				log.Printf("[RATE-LIMIT] Pesan dari %s diabaikan (terlalu cepat)", senderJID)
				return 
			}

			username := v.Info.PushName
			idx := strings.Index(senderJID, "@")
			number := senderJID
			if idx > 0 {
				number = senderJID[:idx]
			}

			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}
			imageMsg := v.Message.GetImageMessage()
			
			replies := botRouter.Route(senderJID, text, imageMsg, v.Info.ID, v.Info.Chat, v.Info.IsGroup, username, number)

			// for _, reply := range replies {
			// 	if reply == "" {
			// 		continue
			// 	}
			// 	SendAsync(context.Background(), botRouter.Ctx.WAClient, v.Info.Chat, reply, "private")
			// }
			go func() {
				for _, reply := range replies {
					if reply != "" {
						// Panggil Common SmartSend
						common.SmartSend(context.Background(), botRouter.Ctx.WAClient, v.Info.Chat, reply)
					}
				}
			}()

		// 2. SOLUSI: Menangani Pesan yang Gagal didekripsi (Undecryptable)
		case *events.UndecryptableMessage:
			log.Printf("[WARN] Menerima pesan yang tidak bisa dibaca dari %s. Mengirim pesan pemulihan...", v.Info.Sender.String())
			
			// Kirim pesan pancingan agar HP user memperbarui kunci enkripsi
			recoveryMsg := "⚠️ *Sistem Pemulihan*\n\nMaaf, server kami baru saja melakukan pembaruan keamanan. Pesan terakhir Anda tidak terbaca.\n\n🔄 *Mohon kirim ulang pesan Anda sekarang.* Terima kasih!"
		
			go common.SmartSend(context.Background(), botRouter.Ctx.WAClient, v.Info.Chat, recoveryMsg)
		}
	}
}
