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
	
	//sesi pengaduan
	s := GetSession(jid)
	if s.Step == "menunggu_pengaduan" {
		return "Mohon kirimkan gambar beserta deskripsi pengaduan Anda."
	}
	
	switch text {
	case "!batal":
		ResetSession(jid)
		return "Sesi Dibatalkan"
		case "2":
		s.Step = "menunggu_pengaduan"
		return "Anda memilih buat pengaduan.\n\nSilakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."
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
