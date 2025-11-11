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

*1.* Data Diri
*2.* Pengajuan Surat 
*3.* Pengaduan 

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
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		subMenu := `*Menu Data Diri*

Gunakan menu ini buat ngatur data kependudukan kamu.

*1.* Input Data Diri (Baru)
*2.* Ubah Data Diri (pakai NIK)

Ketik nomor pilihanmu, atau ketik ‘reset’ buat balik ke menu utama.`
		return []string{subMenu}

	case "2":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_MENU_UTAMA); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_SURAT_MENU_UTAMA: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}

		return []string{
			"*Menu Pengajuan Surat*\n\n"+
			"Silakan pilih jenis surat apa yang ingin anda ajukan: \n"+
			"1. Surat Domisili\n"+
			"2. Surat Usaha\n"+
			"3. Surat Umum\n"+
			"4. Surat Tanggungan\n"+
			"5. Surat Kematian\n"+
			"Ketik nomor surat (1-5) atau 'reset' untuk batal.",
		}

	case "3":
		if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_MENU); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_PENGADUAN_MENU: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		subMenu := `*Menu Pengaduan*
		Menu ini berfungsi untuk mengelola data pengaduan masyarakat.
		1. Ajukan pengaduan
		2. Cek progres pengaduan

		Silakan pilih nomor atau ketik 'reset' untuk kembali ke menu utama.`
			return []string{subMenu}
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
