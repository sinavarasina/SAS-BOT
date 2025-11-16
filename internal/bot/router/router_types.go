package router

import (
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// Router adalah antarmuka untuk router utama bot
type Router interface {
	Route(jid string, text string, imageMsg *waE2E.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string
}
