package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
)

type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

const (
	GEMINI_API_KEY = "AIzaSyCR3Weo_JBPE_PnWNLEfo4T57Uw0bqCQM4" // SEBAIKNYA PINDAHKAN KE .env
	GEMINI_API_URL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
)

// Static responses map
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

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string, sheetsClient *sheets.Data_SheetsClient) []string {
	// Debug raw message content first
	log.Printf("[DEBUG] Raw message - Text: '%s', Length: %d", text, len(text))

	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] After trim - Text: '%s', Length: %d, Username: %s, Number: %s",
		text, len(text), username, number)

	// Handle empty or whitespace-only messages
	if len(strings.TrimSpace(text)) == 0 {
		log.Printf("[DEBUG] Message contains only whitespace or is empty")
		return []string{getMainMenu()}
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
		if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Failed to reset session: %v", err)
			return []string{"Terjadi kesalahan sistem."}
		}
		log.Printf("[DEBUG] Session reset successful")
		return []string{getMainMenu()}
	}

	// Check if we're already in a data entry session
	session, err := db.GetOrCreateDataEntrySession(dbConn, jid)
	if err != nil {
		log.Printf("[ERROR] Session error: %v", err)
		return []string{"Maaf, terjadi kesalahan sistem."}
	}

	// Only process data entry if we're already in a session AND awaiting answer
	if session.AwaitingAnswer {
		log.Printf("[DEBUG] Processing answer for existing session")
		return HandleDataEntry(dbConn, jid, text, session, sheetsClient)
	}

	// --- PERBAIKAN LOGIKA UTAMA ADA DI SINI ---
	// Handle menu selection "1" to start data entry
	if text == "1" {
		// JANGAN panggil StartNewSession di sini.
		// CUKUP set langkah ke menu data diri (200)
		if err := db.UpdateStepOnly(dbConn, jid, STEP_MENU_DATA_DIRI); err != nil {
			log.Printf("[ERROR] Failed to set step to STEP_MENU_DATA_DIRI: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}

		// Tampilkan sub-menu (tanpa "Hapus")
		subMenu := `*Menu Data Diri*

Menu ini digunakan untuk mengelola data kependudukan Anda.

1. Input Data Diri (Baru)
2. Edit Data Diri (Berdasarkan NIK)

Silakan pilih nomor atau ketik 'reset' untuk kembali ke menu utama.`
		return []string{subMenu}
	}
	// --- AKHIR PERBAIKAN ---

	// Check for static responses if not in data entry mode
	if response, exists := getStaticResponse(text); exists {
		return []string{response + "\n\n" + getMainMenu()}
	}

	// Use Gemini API for other messages when not in data entry mode
	response := handleGeminiAPI(text)
	return []string{response + "\n\n" + getMainMenu()}
}

func getStaticResponse(text string) (string, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	response, exists := staticResponses[text]
	return response, exists
}

func handleGeminiAPI(text string) string {
	prompt := fmt.Sprintf("Kamu adalah bot asisten untuk pencatatan data penduduk di desa Sindang Anom. "+
		"Pengguna mengirim: '%s'. "+
		"Berikan respons yang ramah dalam 1-2 kalimat dan arahkan untuk memilih menu 1 untuk pencatatan data. "+
		"Hindari mengulangi teks menu dan jangan menggunakan kata 'Tentu'.", text)

	req := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal request: %v", err)
		return "Maaf, saya tidak mengerti."
	}

	url := fmt.Sprintf("%s?key=%s", GEMINI_API_URL, GEMINI_API_KEY)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] API request failed: %v", err)
		return "Maaf, saya tidak mengerti."
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response: %v", err)
		return "Maaf, saya tidak mengerti."
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("[ERROR] Failed to unmarshal response: %v", err)
		return "Maaf, saya tidak mengerti."
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	}

	return "Maaf, saya tidak mengerti."
}

// Add new function to get main menu text
func getMainMenu() string {
	menu := `*SINDANG ANOM SERVICE - BOT*

Menu yang tersedia:

1. Data Diri
2. Pengajuan Surat 
3. Pengaduan 

Silakan pilih menu dengan mengetik nomor yang sesuai.`
	return menu
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	// Hapus pesan template, bisa return kosong atau pesan singkat lain jika diinginkan
	return ""
}
