package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/whatsapp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appDB, err := db.InitDB("./app_user.db")
	if err != nil {
		log.Fatal("Error at db.InitDB(), Message :", err)
	}

	WaClient, err := whatsapp.InitClient(appDB, ctx)
	if err != nil {
		log.Fatal("Error at whatsapp.InitClient(), Message :", err)
	}

	log.Println("SAS-BOT Running..")

	<-ctx.Done()
	log.Println("Shutting Down SAS-BOT")
	WaClient.Disconnect()
}
