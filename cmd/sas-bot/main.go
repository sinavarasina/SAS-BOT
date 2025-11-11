package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"github.com/sinavarasina/SAS-BOT/internal/whatsapp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Error at loading .env file")
	}
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Printf("[ERROR] Error at os.Getenv('POSTGRES_DSN')")
	}

	appDB, err := db.InitDB(dsn)
	if err != nil {
		log.Fatal("[ERROR] Error at db.InitDB(), Message :", err)
	}

	sheetsClient, err := sheets.InitSheetsClient()
	if err != nil {
		log.Fatal("[ERROR] Error at sheets.InitSheetsClient(), Message :", err)
	}

	driveClient, err := uploader.InitDriveClient()
	if err != nil {
		log.Fatal("[ERROR] Error at uploader.InitDriveClient(), Message :", err)
	}
	WaClient, err := whatsapp.InitClient(dsn, appDB, ctx, sheetsClient, driveClient)
	if err != nil {
		log.Fatal("Error at whatsapp.InitClient(), Message :", err)
	}

	log.Println("SAS-BOT Running..")

	<-ctx.Done()
	log.Println("Shutting Down SAS-BOT")
	WaClient.Disconnect()
}
