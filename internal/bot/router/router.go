package router

import (
	"log"
	"strings"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/bot/gemini"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/datadiri"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/pengaduan"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/surat"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type BotRouter struct {
	Ctx      *common.ServiceContext 
	handlers map[common.FlowName]common.MenuHandler
	gemini   *gemini.GeminiService
}

func NewRouter(ctx *common.ServiceContext, geminiSvc *gemini.GeminiService) *BotRouter {
	handlers := make(map[common.FlowName]common.MenuHandler)
	
	handlers[common.FlowDataDiri] = datadiri.NewHandler(ctx)
	handlers[common.FlowSurat] = surat.NewHandler(ctx)
	handlers[common.FlowPengaduan] = pengaduan.NewHandler(ctx)
	
	log.Println("[INFO] Router berhasil dibuat dengan modul: DataDiri, Surat, Pengaduan")

	return &BotRouter{
		Ctx:      ctx, 
		handlers: handlers,
		gemini:   geminiSvc,
	}
}

func (r *BotRouter) Route(jid string, text string, imageMsg *proto.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string {
	
	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] Menerima input: '%s', JID: %s", text, jid)

	if len(text) == 0 && imageMsg == nil {
		return []string{common.GetMainMenu()}
	}

	go func() {
		user := db.User{JID: jid, Username: username, Number: number}
		if err := db.SaveUser(r.Ctx.DB, user); err != nil {
			log.Printf("[ERROR] Gagal menyimpan user: %v", err)
		}
	}()

	if isGroup {
		if strings.HasPrefix(strings.ToLower(text), "!menu") {
			return []string{"Gunakan private chat untuk mengakses menu layanan. Terima kasih."}
		}
		return nil 
	}

	if common.NormalizeInput(text) == "reset" || common.NormalizeInput(text) == "!batal" {
		if err := db.DeleteDataEntrySession(r.Ctx.DB, jid); err != nil {
			log.Printf("[ERROR] Gagal reset sesi: %v", err)
			return []string{"Terjadi kesalahan sistem."}
		}
		log.Printf("[DEBUG] Sesi direset untuk %s", jid)
		return []string{common.GetMainMenu()}
	}

	session, err := db.GetOrCreateDataEntrySession(r.Ctx.DB, jid)
	if err != nil {
		log.Printf("[ERROR] Gagal mendapatkan sesi: %v", err)
		return []string{"Maaf, terjadi kesalahan sistem."}
	}

	if imageMsg != nil {
		if session.CurrentFlow.String == string(common.FlowPengaduan) && session.CurrentStep == common.STEP_PENGADUAN_WAITING {
			return r.handlers[common.FlowPengaduan].HandleImage(session, imageMsg, messageID, chatJID)
		}
		return []string{"Maaf, saya tidak mengharapkan gambar saat ini."}
	}

	if session.AwaitingAnswer {
		log.Printf("[DEBUG] Merutekan ke Flow: %s, Step: %d", session.CurrentFlow.String, session.CurrentStep)
		
		currentFlow := common.FlowName(session.CurrentFlow.String)
		if handler, ok := r.handlers[currentFlow]; ok {
			return handler.HandleText(session, text)
		}
		
		log.Printf("[WARN] Sesi AwaitingAnswer=true tetapi CurrentFlow=%s tidak dikenal. Mereset sesi.", session.CurrentFlow.String)
		db.DeleteDataEntrySession(r.Ctx.DB, jid) // <-- Gunakan r.Ctx
		return []string{common.GetMainMenu()}
	}

	switch text {
	case "1": 
		if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowDataDiri), common.STEP_MENU_DATA_DIRI); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuDataDiri()}
	
	case "2": 
		if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowSurat), common.STEP_SURAT_MENU_UTAMA); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuSurat()}
		
	case "3": 
		if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowPengaduan), common.STEP_PENGADUAN_MENU); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{common.GetSubmenuPengaduan()}
	}

	response := r.gemini.HandleGeminiPrompt(text)
	return []string{response + "\n\n" + common.GetMainMenu()}
}
