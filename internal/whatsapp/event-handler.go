package whatsapp

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot/router"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// EventHandler sekarang hanya menerima router
func EventHandler(botRouter *router.botRouter, appDB *sqlx.DB) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			
			chatJID := v.Info.Chat.String()
			senderJID := v.Info.Sender.String()
			username := v.Info.PushName
			idx := strings.Index(senderJID, "@")
			number := senderJID
			if idx > 0 {
				number = senderJID[:idx]
			}

			// Ekstrak pesan
			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}
			imageMsg := v.Message.GetImageMessage()
			
			// Panggil Router Utama
			replies := botRouter.Route(senderJID, text, imageMsg, v.Info.ID, v.Info.Chat, v.Info.IsGroup, username, number)

			// Kirim balasan
			for _, reply := range replies {
				if reply == "" {
					continue
				}
				// (Kita gunakan SendAsync dari sender.go)
				SendAsync(context.Background(), botRouter.ctx.WAClient, v.Info.Chat, reply, "private")
			}
		}
	}
}

// (Pindahkan SendAsync dari sender.go ke sini agar lebih mudah)
func SendAsync(ctx context.Context, client *whatsmeow.Client, chat types.JID, text, msgType string) {
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		msg := &waProto.Message{ Conversation: proto.String(text) }
		_, err := client.SendMessage(ctx, chat, msg)
		
		if err != nil {
			log.Printf("[ERROR] Gagal kirim (%s) ke %s: %v", msgType, chat.String(), err)
		} else {
			log.Printf("[SEND] Pesan (%s) terkirim ke %s", msgType, chat.String())
		}
	}()
}
