// file: internal/whatsapp/client.go
package whatsapp

import (
	"context"
	"log"
	"os"

	"github.com/jmoiron/sqlx" // <-- PERBAIKAN 1: Tambahkan import 'sqlx'
	_ "github.com/lib/pq"
	"github.com/mdp/qrterminal"
	"github.com/sinavarasina/SAS-BOT/internal/bot/router"
	// "github.com/sinavarasina/SAS-BOT/internal/db" // <-- PERBAIKAN 2: Hapus import 'db'
	"github.com/sinavarasina/SAS-BOT/internal/sheets"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// NewClient membuat klien WA tetapi TIDAK mendaftarkan handler
func NewClient(dsn string, ctx context.Context) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)

	container, err := sqlstore.New(ctx, "postgres", dsn, dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	return client, nil
}

// InitAndStart mendaftarkan handler DAN menghubungkan klien
// (Tanda tangan fungsi ini sekarang valid karena 'sqlx' sudah di-import)
func InitAndStart(ctx context.Context, client *whatsmeow.Client, appDB *sqlx.DB, sheetsClient *sheets.SheetsClient, botRouter *router.BotRouter) error {
	
	// Daftarkan handler
	client.AddEventHandler(EventHandler(botRouter, appDB))

	// Hubungkan
	if client.Store.ID == nil {
		if err := QRLogin(ctx, client); err != nil {
			return err
		}
	} else {
		if err := client.Connect(); err != nil {
			return err
		}
	}
	return nil
}

func QRLogin(ctx context.Context, client *whatsmeow.Client) error {
	qrChan, _ := client.GetQRChannel(ctx)
	if err := client.Connect(); err != nil {
		return err
	}

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			qrterminal.Generate(evt.Code, qrterminal.L, os.Stdout)
		default:
			log.Println("QR Event:", evt.Event)
		}
	}
	return nil
}
