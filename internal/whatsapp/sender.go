package whatsapp

import (
	"context"
	"log"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SendAsync(ctx context.Context, client *whatsmeow.Client, chat types.JID, msg, label string) {
	go func() {
		msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err := client.SendMessage(msgCtx, chat, &waE2E.Message{
			Conversation: proto.String(msg),
		})
		if err != nil {
			log.Printf("[WARN] Failed to send %s message: %v", label, err)
		} else {
			log.Printf("[SEND] %s message delivered: %s", label, msg)
		}
	}()
}
