package common

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"github.com/sinavarasina/SAS-BOT/internal/utils"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// ServiceContext adalah konteks yang diteruskan ke setiap modul menu
type ServiceContext struct {
	DB            *sqlx.DB
	SheetsClient  *sheets.SheetsClient
	WAClient      *whatsmeow.Client
	R2Uploader    *uploader.R2Uploader
	ImgbbUploader *uploader.ImgbbUploader
	Limiter       *utils.RateLimiter
}

// MenuHandler adalah antarmuka yang harus dipatuhi oleh setiap modul menu
type MenuHandler interface {
	HandleText(session *db.DataEntrySession, text string) []string
	HandleImage(session *db.DataEntrySession, imageMsg *waE2E.ImageMessage, messageID string, chatJID types.JID) []string
}
