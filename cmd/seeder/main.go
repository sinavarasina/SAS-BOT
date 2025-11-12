package main

import (
	"flag"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/cmd/seeder"
)

func main() {
	// Flag untuk custom CSV path
	csvPath := flag.String("csv", "csv/datadiri.csv", "Path ke file CSV")

	// Ambil database URL dari environment variable POSTGRES_DSN
	defaultDBURL := os.Getenv("POSTGRES_DSN")
	if defaultDBURL == "" {
		log.Fatal("[ERROR] Environment variable POSTGRES_DSN tidak di-set di .env")
	}
	dbURL := flag.String("db", defaultDBURL, "Database connection URL")
	flag.Parse()

	log.Printf("[SEEDER] Memulai dengan CSV: %s", *csvPath)
	log.Printf("[SEEDER] Database: %s", *dbURL)

	// Buka koneksi ke database
	dbConn, err := sqlx.Connect("postgres", *dbURL)
	if err != nil {
		log.Fatalf("[ERROR] Gagal koneksi ke database: %v", err)
	}
	defer dbConn.Close()

	// Jalankan seeder
	if err := seeder.SeedDataDariCSV(dbConn, *csvPath); err != nil {
		log.Fatalf("[ERROR] Seeder gagal: %v", err)
	}

	log.Printf("[SEEDER] ✅ Seeder selesai dengan sukses!")
}
