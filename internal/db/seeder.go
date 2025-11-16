package db

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type CSVDataDiri struct {
	Dusun       string
	RT          string
	NoKK        string
	NIK         string
	Nama        string
	TempatLahir string
	TglLahir    string
	KawinID     int64
	SexID       int64
}

// SeedDataDariCSV membaca file CSV dan memasukkan data ke database
func SeedDataDariCSV(dbConn *sqlx.DB, csvPath string) error {
	log.Printf("[SEEDER] Memulai import data dari CSV: %s", csvPath)

	// Buka file CSV
	file, err := os.Open(csvPath)
	if err != nil {
		log.Printf("[ERROR] Gagal membuka file CSV: %v", err)
		return fmt.Errorf("gagal membuka file CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Baca header (baris pertama)
	header, err := reader.Read()
	if err != nil {
		log.Printf("[ERROR] Gagal membaca header CSV: %v", err)
		return fmt.Errorf("gagal membaca header CSV: %v", err)
	}
	log.Printf("[SEEDER] Header: %v", header)

	// Mapping index kolom
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Validasi kolom yang diperlukan
	requiredCols := []string{"dusun", "rt", "no_kk", "nik", "nama", "tempat_lahir", "tgl_lahir", "kawin_id", "sex_id"}
	for _, col := range requiredCols {
		if _, exists := columnMap[col]; !exists {
			return fmt.Errorf("kolom '%s' tidak ditemukan di CSV", col)
		}
	}

	// Baca data baris per baris
	rowCount := 0
	successCount := 0
	failCount := 0

	for {
		record, err := reader.Read()
		if err != nil {
			break // EOF
		}

		rowCount++

		// Parse data dari CSV
		data := CSVDataDiri{
			Dusun:       record[columnMap["dusun"]],
			RT:          record[columnMap["rt"]],
			NoKK:        record[columnMap["no_kk"]],
			NIK:         record[columnMap["nik"]],
			Nama:        record[columnMap["nama"]],
			TempatLahir: record[columnMap["tempat_lahir"]],
			TglLahir:    record[columnMap["tgl_lahir"]],
		}

		// Parse ID fields
		kawinID, err := strconv.ParseInt(record[columnMap["kawin_id"]], 10, 64)
		if err != nil {
			log.Printf("[WARN] Baris %d: kawin_id tidak valid (%s), skip baris ini", rowCount, record[columnMap["kawin_id"]])
			failCount++
			continue
		}
		data.KawinID = kawinID

		sexID, err := strconv.ParseInt(record[columnMap["sex_id"]], 10, 64)
		if err != nil {
			log.Printf("[WARN] Baris %d: sex_id tidak valid (%s), skip baris ini", rowCount, record[columnMap["sex_id"]])
			failCount++
			continue
		}
		data.SexID = sexID

		// Parse tanggal (format DD-MM-YYYY)
		parsedDate, err := time.Parse("02-01-2006", data.TglLahir)
		if err != nil {
			log.Printf("[WARN] Baris %d: format tanggal salah (%s), skip baris ini", rowCount, data.TglLahir)
			failCount++
			continue
		}

		// Insert/Update ke database
		if err := insertOrUpdateDataDiri(dbConn, data, parsedDate); err != nil {
			log.Printf("[ERROR] Baris %d: Gagal insert data: %v", rowCount, err)
			failCount++
			continue
		}

		successCount++
		if successCount%100 == 0 {
			log.Printf("[SEEDER] Progress: %d rows berhasil diproses", successCount)
		}
	}

	log.Printf("[SEEDER] Import selesai. Total baris: %d, Berhasil: %d, Gagal: %d", rowCount-1, successCount, failCount)
	return nil
}

// insertOrUpdateDataDiri insert atau update data ke tabel data_penduduk
func insertOrUpdateDataDiri(dbConn *sqlx.DB, data CSVDataDiri, tanggalLahir time.Time) error {
	query := `
	INSERT INTO data_penduduk (
		nik, no_kk, nama, dusun, rt, tempat_lahir, tanggal_lahir, 
		status_kawin_id, sex_id, created_at, updated_at
	) VALUES (
		:nik, :no_kk, :nama, :dusun, :rt, :tempat_lahir, :tanggal_lahir,
		:status_kawin_id, :sex_id, NOW(), NOW()
	)
	ON CONFLICT (nik) DO UPDATE SET
		no_kk = EXCLUDED.no_kk,
		nama = EXCLUDED.nama,
		dusun = EXCLUDED.dusun,
		rt = EXCLUDED.rt,
		tempat_lahir = EXCLUDED.tempat_lahir,
		tanggal_lahir = EXCLUDED.tanggal_lahir,
		status_kawin_id = EXCLUDED.status_kawin_id,
		sex_id = EXCLUDED.sex_id,
		updated_at = NOW()
	`

	_, err := dbConn.NamedExec(query, map[string]any{
		"nik":             data.NIK,
		"no_kk":           data.NoKK,
		"nama":            data.Nama,
		"dusun":           data.Dusun,
		"rt":              data.RT,
		"tempat_lahir":    data.TempatLahir,
		"tanggal_lahir":   tanggalLahir,
		"status_kawin_id": data.KawinID,
		"sex_id":          data.SexID,
	})

	return err
}

// RunSeeder adalah wrapper function untuk menjalankan seeder dengan path CSV default
func RunSeeder(dbConn *sqlx.DB) error {
	csvPath := filepath.Join("csv", "csv_datadiri.csv")

	// Cek apakah file ada
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		log.Printf("[WARN] File CSV tidak ditemukan di: %s", csvPath)
		return fmt.Errorf("file CSV tidak ditemukan: %s", csvPath)
	}

	return SeedDataDariCSV(dbConn, csvPath)
}
