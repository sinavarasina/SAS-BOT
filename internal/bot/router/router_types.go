// file: internal/bot/router/router_types.go
package router

import (
	// --- TAMBAHKAN 2 IMPORT INI ---
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	// --- AKHIR TAMBAHAN ---
)

// (Struct ServiceContext dihapus dari sini, karena sudah ada di common/types.go)

// Router adalah antarmuka untuk router utama bot
type Router interface {
	// --- PERBAIKAN: Gunakan 'proto.ImageMessage' dan 'types.JID' ---
	Route(jid string, text string, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string
}
