package whatsapp

import (
	"context"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func EventHandler(client *whatsmeow.Client, appDB *sqlx.DB, sheetsClient *sheets.SheetsClient, ctx context.Context) func(interface{}) {
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

			session, err := db.GetOrCreateDataEntrySession(appDB, senderJID)
			if err != nil {
				log.Printf("[ERROR] Gagal mendapatkan sesi untuk %s: %v", senderJID, err)
				return
			}

			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}

			imageMsg := v.Message.GetImageMessage()
			if !v.Info.IsGroup && imageMsg != nil && session.CurrentStep == bot.STEP_PENGADUAN_WAITING {
				go bot.HandleImagePengaduan(
					client,
					appDB,
					sheetsClient,
					senderJID,
					imageMsg,
					v.Info.ID,
					v.Info.Chat,
				)
				return
			}

			if v.Info.IsGroup {
				reply := bot.HandlerRouteGroup(appDB, chatJID, text, username, number)
				if reply != "" {
					SendAsync(ctx, client, v.Info.Chat, reply, "group")
				}
				return
			}

			replies := bot.HandlerRoutePrivate(appDB, chatJID, text, username, number, sheetsClient)
			for _, reply := range replies {
				if reply == "" {
					continue
				}
				select {
				case <-ctx.Done():
					log.Printf("[STOP] Global shutdown detected. Canceling send to %s", chatJID)
					return
				default:
					SendAsync(ctx, client, v.Info.Chat, reply, "private")
				}
			}
		}
	}
}
