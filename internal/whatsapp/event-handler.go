package whatsapp

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot/router"
	"github.com/sinavarasina/SAS-BOT/internal/utils"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func EventHandler(botRouter *router.BotRouter, appDB *sqlx.DB, limiter *utils.RateLimiter) func(any) {
	return func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}

			senderJID := v.Info.Sender.String()

			if !limiter.Allow(senderJID) {
				log.Printf("[RATE-LIMIT] Pesan dari %s diabaikan (terlalu cepat)", senderJID)
				return // Hentikan pemrosesan pesan
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

			// Panggil Router Utama
			replies := botRouter.Route(senderJID, text, imageMsg, v.Info.ID, v.Info.Chat, v.Info.IsGroup, username, number)

			// Kirim balasan
			for _, reply := range replies {
				if reply == "" {
					continue
				}
				SendAsync(context.Background(), botRouter.Ctx.WAClient, v.Info.Chat, reply, "private")
			}
		}
	}
}

// SendAsync mengirim pesan di background
func SendAsync(ctx context.Context, client *whatsmeow.Client, chat types.JID, text, msgType string) {
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		msg := &waE2E.Message{Conversation: proto.String(text)}
		_, err := client.SendMessage(ctx, chat, msg)

		if err != nil {
			log.Printf("[ERROR] Gagal kirim (%s) ke %s: %v", msgType, chat.String(), err)
		} else {
			log.Printf("[SEND] Pesan (%s) terkirim ke %s", msgType, chat.String())
		}
	}()
}
