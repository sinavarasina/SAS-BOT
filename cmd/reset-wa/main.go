package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] Tidak menemukan file .env, mencoba menggunakan environment variables sistem")
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("[FATAL] POSTGRES_DSN tidak ditemukan di .env")
	}

	// 2. Koneksi ke Database
	log.Println("[INFO] Menghubungkan ke database...")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("[FATAL] Gagal koneksi ke database: %v", err)
	}
	defer db.Close()

	// 3. Daftar Tabel yang Akan Dihapus
	queries := []string{
		"DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_history_sync_gen CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_identities CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_prekeys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_device CASCADE;",
	}

	// 4. Eksekusi Penghapusan
	log.Println("[INFO] Memulai pembersihan sesi WhatsApp...")
	
	tx, err := db.Begin() // Gunakan transaksi agar atomik
	if err != nil {
		log.Fatalf("[FATAL] Gagal memulai transaksi: %v", err)
	}

	for _, query := range queries {
		log.Printf("[EXEC] %s", query)
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			log.Fatalf("[FATAL] Gagal mengeksekusi query: %s. Error: %v", query, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("[FATAL] Gagal commit transaksi: %v", err)
	}

	log.Println("-------------------------------------------------------")
	log.Println("[SUCCESS] Sesi WhatsApp BERHASIL DIBERSIHKAN!")
	log.Println("[INFO] Data penduduk dan pengaduan tetap AMAN.")
	log.Println("[INFO] Silakan jalankan bot utama untuk scan QR baru.")
	log.Println("-------------------------------------------------------")
}
