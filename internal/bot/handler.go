package bot

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string) []string {
	log.Printf("[DEBUG] Raw message - Text: '%s', Length: %d", text, len(text))
	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] After trim - Text: '%s', Length: %d, Username: %s, Number: %s",
		text, len(text), username, number)

	if len(strings.TrimSpace(text)) == 0 {
		log.Printf("[DEBUG] Message contains only whitespace or is empty")
		return []string{""}
	}

	// Save user data to database
	WAUser := db.User{
		JID:       jid,
		Number:    number,
		Username:  username,
		Previlege: "user",
	}
	if err := db.SaveUser(dbConn, WAUser); err != nil {
		log.Printf("[ERROR] Failed to save user: %v", err)
	}

	// -----------------------------
	// Handle complaint session
	// -----------------------------
	s := GetSession(jid)
	if s.Step == "menunggu_pengaduan" {
		return []string{"Mohon kirimkan gambar beserta deskripsi pengaduan Anda."}
	}

	// -----------------------------
	// Handle database session
	// -----------------------------
	session, err := db.GetOrCreateDataEntrySession(dbConn, jid)
	if err != nil {
		log.Printf("[ERROR] Failed to get or create session: %v", err)
		return []string{"Terjadi kesalahan sistem."}
	}

	log.Printf("[DEBUG] Current DB session - Step: %d, AwaitingAnswer: %v",
		session.CurrentStep, session.AwaitingAnswer)

	// If user is currently in data entry mode, ignore global commands like "1" or "2"
	if session.AwaitingAnswer || session.CurrentStep > 1 {
		log.Printf("[DEBUG] User is currently in data entry mode. Input '%s' treated as form response, not a global command.", text)
		return HandleDataEntry(dbConn, jid, text, session)
	}

	switch strings.ToLower(text) {

	case "reset", "!batal":
		if err := db.StartNewSession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to reset session: %v", err)
			return []string{"Terjadi kesalahan sistem."}
		}
		ResetSession(jid)
		log.Printf("[DEBUG] Session reset successfully")
		return []string{"Sesi telah direset. Silakan pilih menu. Kirim '1' untuk memulai pendataan."}

	case "1":
		if err := db.StartNewSession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to start new session: %v", err)
			return []string{"Terjadi kesalahan sistem."}
		}
		log.Printf("[DEBUG] Started a new DB session for data entry")
		return []string{steps[1].Question}

	case "2":
		s.Step = "menunggu_pengaduan"
		log.Printf("[DEBUG] Switched to complaint mode for user %s", jid)
		return []string{
			"Anda memilih buat pengaduan.",
			"Silakan kirimkan 1 foto pengaduan Anda, dan tulis deskripsi di caption gambar.",
		}

	default:
		/*
			// Gemini AI handler (i disable it for now)
			topic := "Materi: "
			question := "Pertanyaan: " + text
			geminiResp, err := askGemini(topic, question)
			if err != nil {
				log.Printf("[ERROR] Gemini API error: %v", err)
				return []string{"Maaf, terjadi kesalahan saat memproses permintaan Anda."}
			}
			return []string{geminiResp}
		*/

		log.Printf("[DEBUG] No matching command found. Showing default menu.")
		return []string{"Silakan pilih menu:\n1. Pendataan\n 2. Pengaduan"}
	}
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	return ""
}

