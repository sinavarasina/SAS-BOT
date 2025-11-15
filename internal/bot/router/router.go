package router

import (
	"log"
	"strings"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/bot/gemini"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/datadiri"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/pengaduan"
	"github.comcom/sinavarasina/SAS-BOT/internal/bot/menu/surat"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type botRouter struct {
	ctx      *ServiceContext
	handlers map[common.FlowName]common.MenuHandler
	gemini   *gemini.GeminiService
}

// NewRouter mendaftarkan semua modul menu ke router
func NewRouter(ctx *ServiceContext, geminiSvc *gemini.GeminiService) *botRouter {
	handlers := make(map[common.FlowName]common.MenuHandler)
	handlers[common.FlowDataDiri] = datadiri.NewHandler(ctx)
	handlers[common.FlowSurat] = surat.NewHandler(ctx)
	handlers[common.FlowPengaduan] = pengaduan.NewHandler(ctx)
	
	log.Println("[INFO] Router berhasil dibuat dengan modul: DataDiri, Surat, Pengaduan")

	return &botRouter{
		ctx:      ctx,
		handlers: handlers,
		gemini:   geminiSvc,
	}
}

// Route adalah pengganti HandlerRoutePrivate DAN DataEntryHandler
func (r *botRouter) Route(jid string, text string, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string {
	
	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] Menerima input: '%s', JID: %s", text, jid)

	if len(text) == 0 && imageMsg == nil {
		return []string{common.GetMainMenu()}
	}

	// Simpan info user (async)
	go func() {
		user := db.User{JID: jid, Username: username, Number: number}
		if err := db.SaveUser(r.ctx.DB, user); err != nil {
			log.Printf("[ERROR] Gagal menyimpan user: %v", err)
		}
	}()

	// Handle Grup (logika sederhana)
	if isGroup {
		if strings.HasPrefix(strings.ToLower(text), "!menu") {
			return []string{"Gunakan private chat untuk mengakses menu layanan. Terima kasih."}
		}
		return nil // Jangan balas apa-apa di grup
	}

	// Handle Perintah Global
	if common.NormalizeInput(text) == "reset" || common.NormalizeInput(text) == "!batal" {
		if err := db.DeleteDataEntrySession(r.ctx.DB, jid); err != nil {
			log.Printf("[ERROR] Gagal reset sesi: %v", err)
			return []string{"Terjadi kesalahan sistem."}
		}
		log.Printf("[DEBUG] Sesi direset untuk %s", jid)
		return []string{common.GetMainMenu()}
	}

	// Dapatkan Sesi
	session, err := db.GetOrCreateDataEntrySession(r.ctx.DB, jid)
	if err != nil {
		log.Printf("[ERROR] Gagal mendapatkan sesi: %v", err)
		return []string{"Maaf, terjadi kesalahan sistem."}
	}

	// --- LOGIKA ROUTING UTAMA ---

	// 1. Handle Input Gambar (Hanya untuk Pengaduan)
	if imageMsg != nil {
		if session.CurrentFlow == common.FlowPengaduan && session.CurrentStep == common.STEP_PENGADUAN_WAITING {
			return r.handlers[common.FlowPengaduan].HandleImage(session, imageMsg, messageID, chatJID)
		}
		return []string{"Maaf, saya tidak mengharapkan gambar saat ini."}
	}

	// 2. Handle Input Teks
	// Jika user sedang dalam alur...
	if session.AwaitingAnswer {
		log.Printf("[DEBUG] Merutekan ke Flow: %s, Step: %d", session.CurrentFlow, session.CurrentStep)
		if handler, ok := r.handlers[session.CurrentFlow]; ok {
			return handler.HandleText(session, text)
		}
		// Sesi 'AwaitingAnswer' tapi tidak ada flow? Ini aneh. Reset saja.
		log.Printf("[WARN] Sesi AwaitingAnswer=true tetapi CurrentFlow=%s tidak dikenal. Mereset sesi.", session.CurrentFlow)
		db.DeleteDataEntrySession(r.ctx.DB, jid)
		return []string{common.GetMainMenu()}
	}

	// 3. Handle Menu Utama (jika tidak dalam alur)
	switch text {
	case "1": // Mulai Alur Data Diri
		if err := db.UpdateSessionFlow(r.ctx.DB, jid, common.FlowDataDiri, common.STEP_MENU_DATA_DIRI); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuDataDiri()}
	
	case "2": // Mulai Alur Surat
		if err := db.UpdateSessionFlow(r.ctx.DB, jid, common.FlowSurat, common.STEP_SURAT_MENU_UTAMA); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuSurat()}
		
	case "3": // Mulai Alur Pengaduan
		if err := db.UpdateSessionFlow(r.ctx.DB, jid, common.FlowPengaduan, common.STEP_PENGADUAN_MENU); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuPengaduan()}
	}

	// 4. Handle Gemini (Fallback)
	response := r.gemini.HandleGeminiPrompt(text)
	return []string{response + "\n\n" + common.GetMainMenu()}
}
