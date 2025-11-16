package router

import (
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)


// Router adalah antarmuka untuk router utama bot
type Router interface {
	Route(jid string, text string, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string
}
