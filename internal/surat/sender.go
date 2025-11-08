package surat

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.mau.fi/whatsmeow"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SendMessage(client *whatsmeow.Client, jid string, msg string) error {
	recipient, err := types.ParseJID(jid)

	if err!=nil {
		log.Printf("[Surat-Send] JID tidak valid: %s", jid)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, recipient, &waProto.Message{
		Conversation: proto.String(msg),
	})
	if err != nil {
		log.Printf("[SURAT-SEND] Gagal kirim pesan ke %s: %v", jid, err)
		return err
	}

	log.Printf("[SURAT-SEND] Pesan terkirim ke %s: %s", jid, msg)
	return nil
}

func SendFile(client *whatsmeow.Client, jid string, path string, caption string) error {
	recipient, err := types.ParseJID(jid)

	if err!=nil {
		log.Printf("[Surat-Send] JID tidak valid: %s", jid)
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

	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	uploaded, err := client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		log.Printf("[SURAT-SEND] Upload gagal untuk %s: %v", path, err)
		return err
	}

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:        proto.String(uploaded.URL),
			Mimetype:   proto.String("application/pdf"),
			Title:      proto.String(filepath.Base(path)),
			FileName:   proto.String(filepath.Base(path)),
			FileLength: proto.Uint64(uint64(len(data))),
			MediaKey:   uploaded.MediaKey,
			DirectPath: proto.String(uploaded.DirectPath),
			Caption:    proto.String(caption),
			FileSHA256: uploaded.FileSHA256,
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
