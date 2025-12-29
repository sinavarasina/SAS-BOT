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

func ResolveToPhoneJID(client *whatsmeow.Client, rawJID string) (types.JID, error) {
	// 1. Parse string input menjadi format JID yang benar
	parsed, err := types.ParseJID(rawJID)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("format JID salah: %v", err)
	}

	// 2. Jika ini LID (Multi-Device), kita coba cek apakah ada di database kontak
	if parsed.Server == "lid" {
		contact, err := client.Store.Contacts.GetContact(context.Background(), parsed)
		
		if err == nil && contact.Found {
			return parsed, nil
		}
	}

	return parsed, nil
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

	// Perbesar timeout untuk upload file
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
