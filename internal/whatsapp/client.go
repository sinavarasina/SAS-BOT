package whatsapp

import (
	"context"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitClient(appDB *sqlx.DB, ctx context.Context) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:wa_session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID == nil {
		if err := QRLogin(ctx, client); err != nil {
			return nil, err
		}
	} else {
		if err := client.Connect(); err != nil {
			return nil, err
		}
	}

	return client, nil
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
