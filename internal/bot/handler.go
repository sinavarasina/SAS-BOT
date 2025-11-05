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

var staticResponses = map[string]string{
	"halo":        "Halo! Selamat datang di SAS-BOT. Silakan pilih menu yang tersedia.",
	"hi":          "Hi! Selamat datang di SAS-BOT. Silakan pilih menu yang tersedia.",
	"hello":       "Hello! Selamat datang di SAS-BOT. Silakan pilih menu yang tersedia.",
	"bantuan":     "Untuk bantuan, silakan pilih menu yang tersedia.",
	"help":        "Untuk bantuan, silakan pilih menu yang tersedia.",
	"tolong":      "Saya siap membantu. Silakan pilih menu yang tersedia.",
	"menu":        "Berikut menu yang tersedia.",
	"mulai":       "Untuk memulai, silakan pilih menu yang tersedia.",
	"start":       "Untuk memulai, silakan pilih menu yang tersedia.",
	"selesai":     "Terima kasih telah menggunakan SAS-BOT. Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"done":        "Terima kasih telah menggunakan SAS-BOT. Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"finish":      "Terima kasih telah menggunakan SAS-BOT. Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"thank":       "Sama-sama! Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"thanks":      "Sama-sama! Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"makasih":     "Sama-sama! Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
	"terimakasih": "Sama-sama! Silakan pilih menu yang tersedia jika butuh bantuan lagi.",
}

func getMainMenu() string {
	return `*SINDANG ANOM SERVICE - BOT*

Menu yang tersedia:

1. Data Diri
2. Pengajuan Surat 
3. Pengaduan 

Silakan pilih menu dengan mengetik nomor yang sesuai.`
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
			return []string{"Terjadi kesalahan sistem."}
		}
		return []string{getMainMenu()}
	}

	session, err := db.GetOrCreateDataEntrySession(dbConn, jid)
	if err != nil {
		log.Printf("[ERROR] Session error: %v", err)
		return []string{"Maaf, terjadi kesalahan sistem."}
	}

	// im kinda frustate, so i choose this stupid nuclear way.
	// if session.CurrentStep >= surat.STEP_SURAT_MENU {
	// 	session.AwaitingAnswer = false
	// }

	if session.AwaitingAnswer {
		return HandleDataEntryCompat(dbConn, jid, text, session, sheetsClient)
	}

	// to start the Surat Menu
	// if session.CurrentStep >= surat.STEP_SURAT_MENU {
	// 	return surat.Handle(dbConn, jid, text, session, waClient)
	// }

	switch text {
	case "1":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_MENU_DATA_DIRI); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_MENU_DATA_DIRI: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		subMenu := `*Menu Data Diri*

Menu ini digunakan untuk mengelola data kependudukan Anda.

1. Input Data Diri (Baru)
2. Edit Data Diri (Berdasarkan NIK)

Silakan pilih nomor atau ketik 'reset' untuk kembali ke menu utama.`
		return []string{subMenu}

	// case "2":
	// 	log.Printf("[SURAT] Masuk ke mode pengajuan surat untuk %s", jid)
	// 	if err := db.UpdateStepOnly(dbConn, jid, 0); err != nil {
	// 		log.Printf("[SURAT-DB] Gagal reset step surat: %v", err)
	// 		return []string{"Terjadi kesalahan membuka menu surat."}
	// 	}
	// 	session.CurrentStep = 0
	// 	return surat.Handle(dbConn, jid, "init", session, waClient)

	case "3":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_WAITING); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_PENGADUAN_WAITING: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Anda memilih *Pengaduan*.\n\nSilakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."}
	}

	if resp, ok := staticResponses[strings.ToLower(text)]; ok {
		return []string{resp + "\n\n" + getMainMenu()}
	}

	if !session.AwaitingAnswer && session.CurrentStep < STEP_MENU_DATA_DIRI {
		log.Printf("[LOG] Calling Gemini for step=%d, text=%q", session.CurrentStep, text)
		resp := HandleGeminiPrompt(text)
		return []string{resp + "\n\n" + getMainMenu()}
	}

	return []string{getMainMenu()}
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "!menu") {
		return "Gunakan private chat untuk mengakses menu layanan. Terima kasih."
	}
	return ""
}
