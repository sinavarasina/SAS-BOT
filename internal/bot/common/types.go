package common

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
)

// BotContext berisi semua dependensi (DB, Sheets, WA, Uploader)
type BotContext struct {
	DB           *sqlx.DB
	SheetsClient *sheets.SheetsClient
	WAClient     *whatsmeow.Client
	R2Uploader   *uploader.R2Uploader // (Kita akan buat ini nanti)
	ImgbbUploader *uploader.ImgbbUploader // (Kita akan buat ini nanti)
}

// ServiceContext adalah konteks yang diteruskan ke setiap modul menu
type ServiceContext struct {
	DB           *sqlx.DB
	SheetsClient *sheets.SheetsClient
	WAClient     *whatsmeow.Client
	R2Uploader   *uploader.R2Uploader
	ImgbbUploader *uploader.ImgbbUploader
}

// MenuHandler adalah antarmuka yang harus dipatuhi oleh setiap modul menu
// (DataDiri, Surat, Pengaduan)
type MenuHandler interface {
	// HandleText menangani input teks
	HandleText(session *db.DataEntrySession, text string) []string
	
	// HandleImage menangani input gambar (hanya relevan untuk Pengaduan)
	HandleImage(session *db.DataEntrySession, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID) []string
}
