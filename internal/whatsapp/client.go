package whatsapp

import (
	"context"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mdp/qrterminal"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitClient(dsn string, appDB *sqlx.DB, ctx context.Context, sheetsClient *sheets.SheetsClient, uploaderClient *uploader.UploadeClient) (*whatsmeow.Client, error) {
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

	client.AddEventHandler(EventHandler(client, appDB, sheetsClient, ctx, driveClient))

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
