package router

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"
	// "math/rand"

	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/bot/gemini"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/datadiri"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/pengaduan"
	"github.com/sinavarasina/SAS-BOT/internal/bot/menu/surat"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	googleProto "google.golang.org/protobuf/proto"
)

type BotRouter struct {
	Ctx          *common.ServiceContext
	handlers     map[common.FlowName]common.MenuHandler
	gemini       *gemini.GeminiService
	sessionCache *sync.Map
}

func NewRouter(ctx *common.ServiceContext, geminiSvc *gemini.GeminiService) *BotRouter {
	handlers := make(map[common.FlowName]common.MenuHandler)

	handlers[common.FlowDataDiri] = datadiri.NewHandler(ctx)
	handlers[common.FlowSurat] = surat.NewHandler(ctx)
	handlers[common.FlowPengaduan] = pengaduan.NewHandler(ctx)

	log.Println("[INFO] Router berhasil dibuat dengan modul: DataDiri, Surat, Pengaduan")

	return &BotRouter{
		Ctx:          ctx,
		handlers:     handlers,
		gemini:       geminiSvc,
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

func (r *BotRouter) Route(jid string, text string, imageMsg *waE2E.ImageMessage, messageID string, chatJID types.JID, isGroup bool, username string, number string) []string {

	text = strings.TrimSpace(text)
	log.Printf("[DEBUG] Menerima input: '%s', JID: %s", text, jid)

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

		if err := db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK); err != nil { /*...*/
		}
		return []string{"Sesi Anda telah direset.\n\nSelamat datang! 🤖 Silakan masukkan NIK 16 digit Anda untuk memulai:"}
	}

	go func() {
		user := db.User{JID: jid, Username: username, Number: number}
		if err := db.SaveUser(r.Ctx.DB, user); err != nil {
			log.Printf("[ERROR] Gagal menyimpan user: %v", err)
		}
	}()

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
			return handler.HandleText(session, text)
		}
		log.Printf("[WARN] Sesi AwaitingAnswer=true tetapi CurrentFlow=%s tidak dikenal. Mereset sesi.", session.CurrentFlow.String)
		db.DeleteDataEntrySession(r.Ctx.DB, jid)
		if err := db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK); err != nil { /*...*/
		}
		return []string{"Terjadi kesalahan, sesi Anda telah direset. Silakan masukkan NIK 16 digit Anda untuk memulai:"}
	}

	switch session.CurrentStep {

	case common.STEP_AWAL_WAIT_NIK: // State: 1 (Menunggu NIK)
		// Validasi format NIK
		if len(text) != 16 || !common.IsNumeric(text) {
			return []string{"Format NIK salah. Harap masukkan 16 digit NIK Anda:"}
		}

		nik := text
		dataPenduduk, err := db.GetDataPendudukByNIK(r.Ctx.DB, nik)

		if err != nil {
			// NIK Tidak Ditemukan
			log.Printf("[DEBUG] NIK %s tidak ditemukan: %v", nik, err)
			// Simpan NIK yang gagal untuk opsi "Daftar Baru"
			if err := db.UpdateSessionField(r.Ctx.DB, jid, "surat_temp_answer", nik); err != nil { /*...*/
			}
			// Pindah ke state "NIK Not Found"
			if err := db.UpdateSessionState(r.Ctx.DB, jid, common.STEP_AWAL_NIK_NOT_FOUND, false); err != nil { /*...*/
			}
			return []string{common.GetNikNotFoundMenu(nik)}

		} else {
			//NIK Ditemukan (Login Sukses)
			log.Printf("[DEBUG] NIK %s ditemukan untuk %s", nik, dataPenduduk.Nama.String)
			// Simpan NIK valid
			if err := db.UpdateSessionField(r.Ctx.DB, jid, "surat_valid_nik", nik); err != nil { /*...*/
			}
			// Pindah ke state "Menu Utama"
			if err := db.UpdateSessionState(r.Ctx.DB, jid, common.STEP_AWAL_MENU_UTAMA, false); err != nil { /*...*/
			}
			return []string{"Selamat datang kembali, " + dataPenduduk.Nama.String + "! 👋\n\n" + common.GetMainMenu()}
		}

	case common.STEP_AWAL_NIK_NOT_FOUND: // State: 2
		switch text {
		case "1": // Ulangi NIK
			if err := db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK); err != nil { /*...*/
			}
			return []string{"Baik, silakan masukkan kembali NIK 16 digit Anda:"}

		case "2": // Daftar Baru
			// Ambil NIK yang gagal
			failedNik, err := db.GetSessionField(r.Ctx.DB, jid, "surat_temp_answer")
			if err != nil || !failedNik.Valid {
				if err := db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK); err != nil { /*...*/
				}
				return []string{"Terjadi kesalahan saat mengambil NIK Anda. Silakan masukkan NIK lagi:"}
			}

			// Mulai Sesi Data Diri Baru (mengatur AwaitingAnswer=true)
			if err := db.StartNewSession(r.Ctx.DB, jid); err != nil {
				return []string{"Kesalahan sistem..."}
			}

			// Pra-isi NIK
			if err := db.UpdateDataEntrySession(r.Ctx.DB, jid, "nik", failedNik.String); err != nil { /*...*/
			}

			// Mulai dari STEP_DUSUN (step 1 dari datadiri)
			startStep := datadiri.STEP_DUSUN
			if err := db.UpdateStepOnly(r.Ctx.DB, jid, startStep); err != nil { /*...*/
			}

			return []string{
				"Baik, kita mulai pendaftaran menggunakan NIK `" + failedNik.String + "`.\n\n" +
					datadiri.FormatQuestion(datadiri.Steps[startStep]),
			}

		default:
			failedNik, _ := db.GetSessionField(r.Ctx.DB, jid, "surat_temp_answer")
			return []string{"Pilihan tidak valid. " + common.GetNikNotFoundMenu(failedNik.String)}
		}

	case common.STEP_AWAL_MENU_UTAMA: // State: 3 ("Login", Menunggu 1, 2, 3)
		switch text {
		case "1": // Menu Data Diri
			// Ambil NIK yang sudah tervalidasi
			nik, err := db.GetSessionField(r.Ctx.DB, jid, "surat_valid_nik")
			if err != nil || !nik.Valid {
				db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK)
				return []string{"Sesi NIK Anda hilang. Silakan masukkan NIK Anda lagi:"}
			}

			// Muat data penduduk ke sesi
			data, err := db.GetDataPendudukByNIK(r.Ctx.DB, nik.String)
			if err != nil {
				return []string{"Gagal mengambil data penduduk Anda. Silakan 'reset'."}
			}
			if err := db.LoadSessionFromPenduduk(r.Ctx.DB, jid, *data); err != nil {
				return []string{"Gagal memuat sesi. Silakan 'reset'."}
			}

			dataStr, _ := db.GetFormattedSessionData(r.Ctx.DB, jid)

			// Langsung lompat ke konfirmasi/edit
			if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowDataDiri), common.STEP_KONFIRMASI_DATA_DIRI); err != nil { /*...*/
			}

			return []string{"Berikut adalah data diri Anda yang terdaftar:\n\n" + dataStr, "\n\nKetik 'valid' untuk konfirmasi (jika tidak ada perubahan) atau 'edit' untuk mengubah data."}

		case "2": // Menu Surat
			if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowSurat), common.STEP_SURAT_MENU_UTAMA); err != nil { /*...*/
			}
			return []string{common.GetSubmenuSurat()}

		case "3": // Menu Pengaduan
			if err := db.UpdateSessionFlow(r.Ctx.DB, jid, string(common.FlowPengaduan), common.STEP_PENGADUAN_MENU); err != nil { /*...*/
			}
			return []string{common.GetSubmenuPengaduan()}

		default: // Panggil Gemini (Asinkron)
			replies := []string{"🤖 Saya proses dulu ya... Mohon tunggu sebentar."}
			go func(ctx *common.ServiceContext, chat types.JID, prompt string) {
				geminiResponse := r.gemini.HandleGeminiPrompt(prompt)
				finalReply := geminiResponse + "\n\n" + common.GetMainMenu()
				SendAsync(context.Background(), ctx.WAClient, chat, finalReply, "gemini_reply")
			}(r.Ctx, chatJID, text)
			return replies
		}

	default:
		if err := db.UpdateStepOnly(r.Ctx.DB, jid, common.STEP_AWAL_WAIT_NIK); err != nil { /*...*/
		}
		return []string{"Selamat datang! 🤖 Silakan masukkan NIK 16 digit Anda untuk memulai:"}
	}
}

// SendAsync (Salinan dari event-handler.go)
func SendAsync(ctx context.Context, client *whatsmeow.Client, chat types.JID, text, msgType string) {
	go func() {
		typingSecs := len(text) / 20 
		if typingSecs < 1 {
			typingSecs = 1
		}

		if typingSecs > 3 {
			typingSecs = 3 // Maksimal nunggu 3 detik agar tidak dianggap spam 
		}

		if err := client.SendChatPresence(chat, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
			log.Printf("[WARN] Gagal kirim presence: %v", err)
		}

		time.Sleep(time.Duration(typingSecs) * time.Second)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		msg := &waE2E.Message{Conversation: googleProto.String(text)}
		_, err := client.SendMessage(ctx, chat, msg)
		//status 'typing' to false 
		_ = client.SendChatPresence(chat, types.ChatPresencePaused, types.ChatPresenceMediaText) 

		if err != nil {
			log.Printf("[ERROR] Gagal kirim (%s) ke %s: %v", msgType, chat.String(), err)
		} else {
			log.Printf("[SEND] Pesan (%s) terkirim ke %s", msgType, chat.String())
		}
	}()
}
