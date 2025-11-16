package surat

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"

	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SendMessage(client *whatsmeow.Client, jid string, msg string) error {
	cleanJID := CleanUserJID(jid)

	recipient, err := types.ParseJID(cleanJID)
	if err != nil {
		log.Printf("[SURAT-SEND] JID tidak valid: %s", cleanJID)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, recipient, &waE2E.Message{
		Conversation: proto.String(msg),
	})
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal kirim pesan ke %s: %v", jid, err)
		return err
	}
	return nil
}

func SendFile(client *whatsmeow.Client, jid string, path string, caption string) error {
	cleanJID := CleanUserJID(jid)

	recipient, err := types.ParseJID(cleanJID)
	if err != nil {
		log.Printf("[SURAT-SEND] JID tidak valid: %s", cleanJID)
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	uploaded, err := client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		log.Printf("[SURAT-SEND] Upload gagal untuk %s: %v", path, err)
		return err
	}

	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String("application/pdf"),
			Title:         proto.String(filepath.Base(path)),
			FileName:      proto.String(filepath.Base(path)),
			FileLength:    proto.Uint64(uint64(len(data))),
			MediaKey:      uploaded.MediaKey,
			DirectPath:    proto.String(uploaded.DirectPath),
			Caption:       proto.String(caption),
			FileSHA256:    uploaded.FileSHA256,
			FileEncSHA256: uploaded.FileEncSHA256,
		},
	}
	_, err = client.SendMessage(ctx, recipient, msg)
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal kirim file ke %s: %v", jid, err)
		return err
	}
	log.Printf("[SURAT-SEND] File terkirim ke %s: %s", jid, path)
	return nil
}

func CleanUserJID(jid string) string {
	parts := strings.Split(jid, "@")
	user := parts[0]

	if strings.Contains(user, ":") {
		user = strings.Split(user, ":")[0]
	}

	return user + "@s.whatsapp.net"
}
