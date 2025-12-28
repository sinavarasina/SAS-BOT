package common

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	googleProto "google.golang.org/protobuf/proto"
)

var sendLock sync.Mutex

// SmartSend: Pengirim pesan cerdas dengan antrian & simulasi manusia
func SmartSend(ctx context.Context, client *whatsmeow.Client, chat types.JID, text string) {
	if text == "" {
		return
	}

	// 1. KUNCI PINTU (Hanya 1 pesan boleh diproses dalam satu waktu di seluruh bot)
	sendLock.Lock()
	defer sendLock.Unlock()

	// 2. Hitung waktu "mengetik" (Humanizing)
	// Rumus: 40ms per karakter. Min 1 detik, Max 4 detik.
	charCount := len(text)
	typingDelay := time.Duration(charCount*40) * time.Millisecond
	if typingDelay < 1*time.Second {
		typingDelay = 1 * time.Second
	}
	if typingDelay > 4*time.Second {
		typingDelay = 4 * time.Second
	}

	// 3. Tambahkan Jitter (Variasi Acak +/- 500ms) agar tidak pola robot
	// Random source perlu seed biar beda tiap restart, tapi sederhana saja:
	jitter := time.Duration(rand.Intn(1000)-500) * time.Millisecond
	finalDelay := typingDelay + jitter
	if finalDelay < 500*time.Millisecond {
		finalDelay = 500 * time.Millisecond
	}

	// 4. Kirim status "Typing..."
	if client != nil {
		_ = client.SendChatPresence(chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	}

	// 5. Tidur sesuai waktu mengetik
	select {
	case <-time.After(finalDelay):
	case <-ctx.Done():
		return
	}

	// 6. Kirim Pesan
	msg := &waE2E.Message{Conversation: googleProto.String(text)}
	if client != nil {
		_, err := client.SendMessage(ctx, chat, msg)
		// Stop typing status
		_ = client.SendChatPresence(chat, types.ChatPresencePaused, types.ChatPresenceMediaText)

		if err != nil {
			log.Printf("[ERROR] Gagal kirim ke %s: %v", chat.String(), err)
		} else {
			log.Printf("[SENT] Pesan terkirim ke %s (Delay: %v)", chat.String(), finalDelay)
		}
	}

	// 7. COOLDOWN (Penting!)
	// Istirahat sebentar sebelum melepas kunci untuk pesan berikutnya
	time.Sleep(time.Duration(rand.Intn(500)+300) * time.Millisecond)
}
