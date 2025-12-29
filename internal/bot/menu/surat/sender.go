package surat

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

//pengganti cleanJid 
func ResolveToPhoneJID(client *whatsmeow.Client, rawJID string) (types.JID, error) {
	// 1. Parse string input menjadi format JID yang benar
	parsed, err := types.ParseJID(rawJID)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("format JID salah: %v", err)
	}

	// 2. Jika bukan LID (sudah format @s.whatsapp.net atau @g.us), pakai langsung
	if parsed.Server != "lid" {
		return parsed, nil
	}

	// 3. Jika LID, cari data aslinya di Database (Store) bot
	contact, err := client.Store.Contacts.GetContact(parsed)
	
	// 4. Jika ketemu, kembalikan JID asli (Nomor HP)
	if err == nil && contact.Found {
		log.Printf("[JID-RESOLVE] Sukses convert LID %s -> Phone %s", parsed.User, contact.JID.User)
		return contact.JID, nil
	}

	// 5. Jika tidak ketemu di database (kasus langka), return error
	return parsed, fmt.Errorf("gagal menemukan nomor HP untuk LID ini (user mungkin belum pernah chat bot)")
}

func SendMessage(client *whatsmeow.Client, jid string, msg string) error {
	recipient, err := ResolveToPhoneJID(client, jid)
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal resolve JID %s: %v", jid, err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, recipient, &waE2E.Message{
		Conversation: proto.String(msg),
	})
	
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal kirim teks ke %s: %v", recipient, err)
		return err
	}
	return nil
}

func SendFile(client *whatsmeow.Client, jid string, path string, caption string) error {
	recipient, err := ResolveToPhoneJID(client, jid)
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal resolve JID %s untuk file: %v", jid, err)
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
		log.Printf("[SURAT-SEND] Gagal kirim file ke %s: %v", recipient, err)
		return err
	}
	
	log.Printf("[SURAT-SEND] File terkirim ke %s (Asal: %s)", recipient, jid)
	return nil
}
