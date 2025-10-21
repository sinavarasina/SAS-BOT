package bot

import (
	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"log"
	"strings"
)

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string) string {
	text = strings.TrimSpace(text)

	WAUser := db.User{
		JID:       jid,
		Number:    number,
		Username:  username,
		Previlege: "user",
	}

	err := db.SaveUser(dbConn, WAUser)
	if err != nil {
		log.Printf("Error at db.SaveUser for jid: %s, Message : %v", jid, err)
	}

	switch text {
	case "!batal":
		ResetSession(jid)
		return "Sesi Dibatalkan"
	}

	topic := "Materi: "
	question := "Pertanyaan: " + text
	geminiResp, err := askGemini(topic, question)
	if err != nil {
		log.Printf("Error Gemini API: %v", err)
		return "Maaf, terjadi kesalahan saat memproses permintaan Anda."
	}
	return geminiResp
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	return ""
}
