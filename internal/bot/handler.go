package bot

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	// "github.com/sinavarasina/SAS-BOT/internal/surat"
	"go.mau.fi/whatsmeow"
)



func getMainMenu() string {
	return `🤖 *SINDANG ANOM SERVICE - BOT*

📋 _Menu yang tersedia:_

👤 *1. Data Diri*
📄 *2. Pengajuan Surat* 
💬 *3. Pengaduan* 

✍️ _Ketik *nomor menu* yang ingin kamu pilih._`
}

func getDataDiriMenu() string {
	return `👤 *Menu Data Diri*

📝 _Gunakan menu ini untuk mengatur data kependudukan kamu._

✏️ *1. Input Data Diri (Baru)*
🔄 *2. Ubah Data Diri (pakai NIK)*

⌨️ _Ketik *nomor (1-2)*, atau ketik *'reset'* untuk kembali ke menu utama._`
}

func getSuratMenu() string {
	return `📄 *Menu Pengajuan Surat*

📝 _Gunakan menu ini untuk mengatur pengajuan surat kamu._

✏️ *1. Ajukan Surat Baru*
🔍 *2. Cek Progres Surat*

⌨️ _Ketik *nomor (1-2)* atau ketik *'reset'* untuk kembali ke menu utama._`
}

func getPengaduanMenu() string {
	return `📢 *Menu Pengaduan*

📝 _Gunakan menu ini untuk mengelola data pengaduan masyarakat._

✏️ *1. Ajukan pengaduan*
🔍 *2. Cek progres pengaduan*

⌨️ _Ketik *nomor (1-2)* atau ketik *'reset'* untuk kembali ke menu utama._`
}

func getSystemError() string {
	return `❌ Maaf, terjadi kesalahan sistem.

Silakan coba lagi atau ketik *'reset'* untuk kembali ke menu utama.`
}

func getGroupChatMessage() string {
	return `📱 Gunakan private chat untuk mengakses menu layanan. 

💬 Silakan chat bot secara langsung untuk menggunakan layanan.`
}

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client) []string {
	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] After trim - Text: '%s', Length: %d, Username: %s, Number: %s", text, len(text), username, number)

	if len(text) == 0 {
		return []string{getMainMenu()}
	}

	user := db.User{JID: jid, Username: username, Number: number}
	if err := db.SaveUser(dbConn, user); err != nil {
		log.Printf("[ERROR] Failed to save user: %v", err)
	}

	switch strings.ToLower(text) {
	case "reset", "!batal":
		if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to reset session: %v", err)
			return []string{getSystemError()}
		}
		return []string{getMainMenu()}
	}

	session, err := db.GetOrCreateDataEntrySession(dbConn, jid)
	if err != nil {
		log.Printf("[ERROR] Session error: %v", err)
		return []string{getSystemError()}
	}

	// im kinda frustate, so i choose this stupid nuclear way.
	// if session.CurrentStep >= surat.STEP_SURAT_MENU {
	// 	session.AwaitingAnswer = false
	// }

	if session.AwaitingAnswer {
		log.Printf("[DEBUG] Processing answer for existing session")
		return HandleDataEntryCompat(dbConn, jid, text, session, sheetsClient, waClient)
	}

	// to start the Surat Menu
	// if session.CurrentStep >= surat.STEP_SURAT_MENU {
	// 	return surat.Handle(dbConn, jid, text, session, waClient)
	// }

	switch text {
	case "1":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_MENU_DATA_DIRI); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_MENU_DATA_DIRI: %v", err)
			return []string{getSystemError()}
		}
		return []string{getDataDiriMenu()}

	case "2":
		// PERBAIKAN: Set ke Menu Utama Surat (500)
		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_MENU_UTAMA); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_SURAT_MENU_UTAMA: %v", err)
			return []string{getSystemError()}
		}
		return []string{getSuratMenu()}

	case "3":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_MENU); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_PENGADUAN_MENU: %v", err)
			return []string{getSystemError()}
		}
		return []string{getPengaduanMenu()}
	}

	if !session.AwaitingAnswer && session.CurrentStep < STEP_MENU_DATA_DIRI {
		go log.Printf("[LOG] Calling Gemini for step=%d, text=%q", session.CurrentStep, text)
		
		// Get menu first (instant, no blocking)
		menu := getMainMenu()
		
		// Get Gemini response (this is the only blocking part)
		resp := HandleGeminiPrompt(text)
		
		// Concatenate in memory (very fast, non-blocking)
		return []string{resp + "\n\n" + menu}
	}

	return []string{getMainMenu()}
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "!menu") {
		return getGroupChatMessage()
	}
	return ""
}