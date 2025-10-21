package bot

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string) string {
	// Debug raw message content first
	log.Printf("[DEBUG] Raw message - Text: '%s', Length: %d", text, len(text))

	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] After trim - Text: '%s', Length: %d, Username: %s, Number: %s",
		text, len(text), username, number)

	// Handle empty or whitespace-only messages
	if len(strings.TrimSpace(text)) == 0 {
		log.Printf("[DEBUG] Message contains only whitespace or is empty")
		return ""
	}

	// Save user info
	WAUser := db.User{
		JID:       jid,
		Number:    number,
		Username:  username,
		Previlege: "user",
	}
	if err := db.SaveUser(dbConn, WAUser); err != nil {
		log.Printf("[ERROR] Failed to save user: %v", err)
	}

	// Handle reset command first
	if strings.ToLower(text) == "reset" || strings.ToLower(text) == "!batal" {
		if err := db.StartNewSession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to reset session: %v", err)
			return "Terjadi kesalahan sistem."
		}
		log.Printf("[DEBUG] Session reset successful")
		return "Sesi telah direset. Silakan pilih menu. Kirim '1' untuk memulai pendataan."
	}

	// Get current session
	session, err := db.GetOrCreateDataEntrySession(dbConn, jid)
	if err != nil {
		log.Printf("[ERROR] Session error: %v", err)
		return "Terjadi kesalahan sistem."
	}

	log.Printf("[DEBUG] Current session state - Step: %d, Awaiting: %v",
		session.CurrentStep, session.AwaitingAnswer)

	// Start new session if user sends "1"
	if text == "1" {
		if err := db.StartNewSession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to start new session: %v", err)
			return "Terjadi kesalahan sistem."
		}
		log.Printf("[DEBUG] Started new session, sending first question")
		return steps[1].Question
	}

	// Handle ongoing session
	if session.AwaitingAnswer {
		log.Printf("[DEBUG] Processing answer for step %d", session.CurrentStep)
		return HandleDataEntry(dbConn, jid, text, session)
	}

	log.Printf("[DEBUG] No active session, showing menu")
	return "Silakan pilih menu. Kirim '1' untuk memulai pendataan."
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	// Hapus pesan template, bisa return kosong atau pesan singkat lain jika diinginkan
	return ""
}
