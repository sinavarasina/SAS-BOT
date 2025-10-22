package whatsapp

import (
	"context"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func EventHandler(client *whatsmeow.Client, appDB *sqlx.DB) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}

			chatJID := v.Info.Chat.String()
			senderJID := v.Info.Sender.String()
			text := v.Message.GetConversation()
			if text == "" && v.Message.ExtendedTextMessage != nil {
				text = v.Message.ExtendedTextMessage.GetText()
			}

			username := v.Info.PushName
			number := senderJID[:strings.Index(senderJID, "@")]

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
