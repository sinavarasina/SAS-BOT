package bot

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"go.mau.fi/whatsmeow"
)

// fungsi abstraksi (untuk layer kompatibilitas sementara <Sampai ketika data entry di pecah menjadi module>)

func HandleDataEntryCompat(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client, driveClient *uploader.DriveClient) []string {
	return HandleDataEntry(dbConn, jid, text, session, sheetsClient, waClient, driveClient)
}
