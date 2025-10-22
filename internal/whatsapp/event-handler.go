package whatsapp

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
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
			number := senderJID[:strings.Index(senderJID, "@")]

			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}

			s := bot.GetSession(senderJID)
			imageMsg := v.Message.GetImageMessage()
			if !v.Info.IsGroup && imageMsg != nil && s.Step == "menunggu_pengaduan" {
				bot.HandleImagePengaduan(client, appDB, sheetsClient, senderJID, imageMsg, v.Info.ID, v.Info.Chat)
				return
			}

			if v.Info.IsGroup {
				reply := bot.HandlerRouteGroup(appDB, chatJID, text, username, number)
				if reply != "" {
					go func(msg string) {
						msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
						defer cancel()

						select {
						case <-ctx.Done():
							log.Printf("[STOP] Global shutdown detected. Canceling send to group %s", chatJID)
							return
						default:
						}

						_, err := client.SendMessage(msgCtx, v.Info.Chat, &waProto.Message{
							Conversation: proto.String(msg),
						})
						if err != nil {
							log.Printf("[WARN] Failed to send group reply to %s: %v", chatJID, err)
						} else {
							log.Printf("[SEND] Sent group reply to %s: %s", chatJID, msg)
						}
					}(reply)
				}
				return
			}

			replies := bot.HandlerRoutePrivate(appDB, chatJID, text, username, number)
			log.Printf("[DEBUG] Handler returned %d replies for %s", len(replies), senderJID)

			for _, reply := range replies {
				if reply == "" {
					continue
				}

				msgCopy := reply

				go func(msg string) {
					msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
					defer cancel()

					select {
					case <-ctx.Done():
						log.Printf("[STOP] Global shutdown detected. Canceling send to %s", chatJID)
						return
					default:
					}

					_, err := client.SendMessage(msgCtx, v.Info.Chat, &waProto.Message{
						Conversation: proto.String(msg),
					})
					if err != nil {
						log.Printf("[WARN] Failed to send private reply to %s: %v", chatJID, err)
					} else {
						log.Printf("[SEND] Delivered message to %s: %s", chatJID, msg)
					}
				}(msgCopy)
			}
		}
	}
}

