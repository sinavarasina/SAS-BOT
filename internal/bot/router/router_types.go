package router

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
)

// ServiceContext berisi semua dependensi yang dibutuhkan oleh modul menu
type ServiceContext struct {
	DB            *sqlx.DB
	SheetsClient  *sheets.SheetsClient
	WAClient      *whatsmeow.Client
	R2Uploader    *uploader.R2Uploader
	ImgbbUploader *uploader.ImgbbUploader
}

// Router adalah antarmuka untuk router utama bot
type Router interface {
	Route(jid string, text string, imageMsg *whatsmeow.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string
}
