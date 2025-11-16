package router

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/bot/gemini"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/datadiri"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/pengaduan"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/surat"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	googleProto "google.golang.org/protobuf/proto"
)

type BotRouter struct {
	Ctx      *common.ServiceContext
	handlers map[common.FlowName]common.MenuHandler
	gemini   *gemini.GeminiService
	sessionCache *sync.Map
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
		sessionCache: new(sync.Map),
	}
}

// updateCache mengambil data sesi terbaru dari DB dan menyimpannya ke cache.
func (r *BotRouter) updateCache(jid string) {
	// Ambil sesi terbaru dari DB
	updatedSession, err := db.GetSessionByJID(r.Ctx.DB, jid)

	if err != nil {
		// Jika sesi tidak ditemukan (misalnya setelah direset), hapus dari cache
		if err == sql.ErrNoRows {
			r.sessionCache.Delete(jid)
			log.Printf("[CACHE] DELETED %s (not found in DB)", jid)
		} else {
			// Jika ada error lain, jangan ubah cache
			log.Printf("[CACHE] FAILED to update %s: %v", jid, err)
		}
		return
	}

	// Jika berhasil, simpan sesi baru ke cache
	r.sessionCache.Store(jid, updatedSession)
	log.Printf("[CACHE] UPDATED %s (Step: %d, Flow: %s)", jid, updatedSession.CurrentStep, updatedSession.CurrentFlow.String)
}

func (r *BotRouter) Route(jid string, text string, imageMsg *waProto.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string {

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
		r.sessionCache.Delete(jid)
		log.Printf("[CACHE] DELETED %s (Reset)", jid)

		log.Printf("[DEBUG] Sesi direset untuk %s", jid)
		return []string{common.GetMainMenu()}
	}

	var session *db.DataEntrySession
	var err error

	cachedSession, ok := r.sessionCache.Load(jid)
	if ok {
		// Cache Hit
		session = cachedSession.(*db.DataEntrySession)
		log.Printf("[CACHE] HIT for %s (Step: %d)", jid, session.CurrentStep)
	} else {
		// Cache Miss
		log.Printf("[CACHE] MISS for %s. Fetching from DB...", jid)
		session, err = db.GetOrCreateDataEntrySession(r.Ctx.DB, jid)
		if err != nil {
			log.Printf("[ERROR] Gagal mendapatkan sesi: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		// Simpan sesi yang baru diambil ke cache
		r.sessionCache.Store(jid, session)
		log.Printf("[CACHE] STORED %s after miss", jid)
	}
	defer r.updateCache(jid)

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
			// Handler akan memodifikasi DB. 'defer' akan menyegarkan cache setelah ini.
			return handler.HandleText(session, text)
		}

		log.Printf("[WARN] Sesi AwaitingAnswer=true tetapi CurrentFlow=%s tidak dikenal. Mereset sesi.", session.CurrentFlow.String)
		db.DeleteDataEntrySession(r.Ctx.DB, jid)
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

	replies := []string{"🤖 Saya proses dulu ya... Mohon tunggu sebentar."}

	go func(ctx *common.ServiceContext, chat types.JID, prompt string) {
		geminiResponse := r.gemini.HandleGeminiPrompt(prompt)
		finalReply := geminiResponse + "\n\n" + common.GetMainMenu()
		SendAsync(context.Background(), ctx.WAClient, chat, finalReply, "gemini_reply")
	}(r.Ctx, chatJID, text)

	return replies
}

// SendAsync (Salinan dari event-handler.go)
func SendAsync(ctx context.Context, client *whatsmeow.Client, chat types.JID, text, msgType string) {
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		msg := &waProto.Message{Conversation: googleProto.String(text)}
		_, err := client.SendMessage(ctx, chat, msg)

		if err != nil {
			log.Printf("[ERROR] Gagal kirim (%s) ke %s: %v", msgType, chat.String(), err)
		} else {
			log.Printf("[SEND] Pesan (%s) terkirim ke %s", msgType, chat.String())
		}
	}()
}
