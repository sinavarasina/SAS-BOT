package main

import (
	"flag"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

func main() {
	// Flag untuk custom CSV path
	csvPath := flag.String("csv", "csv/csv_datadiri.csv", "Path ke file CSV")
	dbURL := flag.String("db", "postgres://avnadmin:AVNS_s9kSgAdLY8nQqi4ATjz@sas-bot-postgre-db-student-5e9a.j.aivencloud.com:21599/defaultdb?sslmode=require", "Database connection URL")
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
	if err := db.SeedDataDariCSV(dbConn, *csvPath); err != nil {
		log.Fatalf("[ERROR] Seeder gagal: %v", err)
	}

	log.Printf("[SEEDER] ✅ Seeder selesai dengan sukses!")
}
