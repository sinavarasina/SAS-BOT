package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common" 
	"github.com/sinavarasina/SAS-BOT/internal/bot/gemini"
	"github.com/sinavarasina/SAS-BOT/internal/bot/router"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/uploader"
	"github.com/sinavarasina/SAS-BOT/internal/utils"
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

	// 1. Inisiasi semua Klien Eksternal
	appDB, err := db.InitDB(dsn)
	if err != nil {
		log.Fatal("[ERROR] Error at db.InitDB(), Message :", err)
	}

	sheetsClient, err := sheets.InitSheetsClient()
	if err != nil {
		log.Fatal("[ERROR] Error at sheets.InitSheetsClient(), Message :", err)
	}

	// 2. Inisiasi Klien Internal (Uploader & Gemini)
	r2Uploader := uploader.NewR2Uploader(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
		os.Getenv("R2_BUCKET_NAME"),
		os.Getenv("R2_PUBLIC_URL"),
	)

	limiter := utils.NewRateLimiter(3, 10*time.Second)
	log.Println("[INFO] Rate Limiter (5 pesan / 10 detik) diaktifkan.")
	imgbbUploader := uploader.NewImgbbUploader(os.Getenv("IMGBB_API_KEY"))
	geminiService := gemini.NewGeminiService(os.Getenv("GEMINI_API_KEY"))

	// 3. Inisiasi Klien WA
	waClient, err := whatsapp.NewClient(dsn, ctx)
	if err != nil {
		log.Fatal("Error at whatsapp.NewClient(), Message :", err)
	}

	// 4. Buat Konteks Layanan (Service Context)
	serviceCtx := &common.ServiceContext{
		DB:            appDB,
		SheetsClient:  sheetsClient,
		WAClient:      waClient,
		R2Uploader:    r2Uploader,
		ImgbbUploader: imgbbUploader,
		Limiter:      limiter,
	}

	// 5. Buat Router Utama
	botRouter := router.NewRouter(serviceCtx, geminiService)

	// 6. Daftarkan Event Handler dan Mulai Koneksi
	if err := whatsapp.InitAndStart(ctx, waClient, appDB, sheetsClient, botRouter, limiter); err != nil {
		log.Fatal("Error at whatsapp.InitAndStart(), Message :", err)
	}
	
	log.Println("SAS-BOT Running..")
	<-ctx.Done()
	log.Println("Shutting Down SAS-BOT")
	waClient.Disconnect()
}
