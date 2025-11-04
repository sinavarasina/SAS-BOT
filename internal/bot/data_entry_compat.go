package bot

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
)

func HandleDataEntryCompat(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient) []string {
	return HandleDataEntry(dbConn, jid, text, session, sheetsClient)
}
