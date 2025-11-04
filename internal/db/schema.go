package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sqlx.DB, error) {
	// Gunakan PostgreSQL driver
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Tes koneksi
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Atur connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// --- SCHEMA CREATION ---
	schema := `
	-- Drop triggers dulu agar tidak konflik
	DROP TRIGGER IF EXISTS update_timestamp ON data_entry_sessions;
	DROP FUNCTION IF EXISTS update_timestamp CASCADE;

	-- Tabel users
	CREATE TABLE IF NOT EXISTS users (
		jid TEXT PRIMARY KEY,
		number TEXT,
		username TEXT,
		previlege TEXT
	);

	-- Tabel pengaduan
	CREATE TABLE IF NOT EXISTS pengaduan (
		id SERIAL PRIMARY KEY,
		user_jid TEXT REFERENCES users(jid),
		deskripsi TEXT,
		pict_path TEXT,
		sent_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Lookup Tables
	CREATE TABLE IF NOT EXISTS sex (sex_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS agama (agama_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS pendidikan_kk (pendidikan_kk_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS pendidikan_sedang (pendidikan_sedang_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS pekerjaan (pekerjaan_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS status_kawin (status_kawin_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS kk_level (kk_level_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS warganegara (warganegara_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS golongan_darah (golongan_darah_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS cacat (cacat_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS cara_kb (cara_kb_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS hamil (hamil_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS ktp_el (ktp_el_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS status_rekam (status_rekam_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS status_dasar (status_dasar_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS suku (suku_id INTEGER PRIMARY KEY, nama TEXT);
	CREATE TABLE IF NOT EXISTS id_asuransi (id_asuransi_id INTEGER PRIMARY KEY, nama TEXT);

	-- Tabel data utama untuk pendataan
	CREATE TABLE IF NOT EXISTS data_entry_sessions (
		jid TEXT PRIMARY KEY,
		current_step INTEGER DEFAULT 1,
		awaiting_answer BOOLEAN DEFAULT false,
		sheet_row_num INTEGER,
		alamat TEXT,
		dusun TEXT,
		rw TEXT,
		rt TEXT,
		nama TEXT,
		no_kk TEXT,
		nik TEXT,
		sex_id INTEGER REFERENCES sex(sex_id),
		tempat_lahir TEXT,
		tanggal_lahir DATE,
		agama_id INTEGER REFERENCES agama(agama_id),
		pendidikan_kk_id INTEGER REFERENCES pendidikan_kk(pendidikan_kk_id),
		pendidikan_sedang_id INTEGER REFERENCES pendidikan_sedang(pendidikan_sedang_id),
		pekerjaan_id INTEGER REFERENCES pekerjaan(pekerjaan_id),
		status_kawin_id INTEGER REFERENCES status_kawin(status_kawin_id),
		kk_level_id INTEGER REFERENCES kk_level(kk_level_id),
		warganegara_id INTEGER REFERENCES warganegara(warganegara_id),
		nik_ayah TEXT,
		nama_ayah TEXT,
		nik_ibu TEXT,
		nama_ibu TEXT,
		golongan_darah_id INTEGER REFERENCES golongan_darah(golongan_darah_id),
		akta_lahir TEXT,
		dokumen_passport TEXT,
		tanggal_akhir_passport DATE,
		dokumen_kitas TEXT,
		akta_perkawinan TEXT,
		tanggal_perkawinan DATE,
		akta_perceraian TEXT,
		tanggal_perceraian DATE,
		cacat_id INTEGER REFERENCES cacat(cacat_id),
		cara_kb_id INTEGER REFERENCES cara_kb(cara_kb_id),
		hamil_id INTEGER REFERENCES hamil(hamil_id),
		ktp_el_id INTEGER REFERENCES ktp_el(ktp_el_id),
		status_rekam_id INTEGER REFERENCES status_rekam(status_rekam_id),
		alamat_sekarang TEXT,
		status_dasar_id INTEGER REFERENCES status_dasar(status_dasar_id),
		suku_id INTEGER REFERENCES suku(suku_id),
		tag_card TEXT,
		id_asuransi_id INTEGER REFERENCES id_asuransi(id_asuransi_id),
		no_asuransi TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS data_penduduk (
		jid TEXT, -- JID dari user yang terakhir mengedit
		nik TEXT PRIMARY KEY, -- NIK sebagai ID Unik
		no_kk TEXT,
		nama TEXT,
		alamat TEXT,
		dusun TEXT,
		rw TEXT,
		rt TEXT,
		sex_id INTEGER REFERENCES sex(sex_id),
		tempat_lahir TEXT,
		tanggal_lahir DATE,
		agama_id INTEGER REFERENCES agama(agama_id),
		pendidikan_kk_id INTEGER REFERENCES pendidikan_kk(pendidikan_kk_id),
		pendidikan_sedang_id INTEGER REFERENCES pendidikan_sedang(pendidikan_sedang_id),
		pekerjaan_id INTEGER REFERENCES pekerjaan(pekerjaan_id),
		status_kawin_id INTEGER REFERENCES status_kawin(status_kawin_id),
		kk_level_id INTEGER REFERENCES kk_level(kk_level_id),
		warganegara_id INTEGER REFERENCES warganegara(warganegara_id),
		nik_ayah TEXT,
		nama_ayah TEXT,
		nik_ibu TEXT,
		nama_ibu TEXT,
		golongan_darah_id INTEGER REFERENCES golongan_darah(golongan_darah_id),
		akta_lahir TEXT,
		dokumen_passport TEXT,
		tanggal_akhir_passport DATE,
		dokumen_kitas TEXT,
		akta_perkawinan TEXT,
		tanggal_perkawinan DATE,
		akta_perceraian TEXT,
		tanggal_perceraian DATE,
		cacat_id INTEGER REFERENCES cacat(cacat_id),
		cara_kb_id INTEGER REFERENCES cara_kb(cara_kb_id),
		hamil_id INTEGER REFERENCES hamil(hamil_id),
		ktp_el_id INTEGER REFERENCES ktp_el(ktp_el_id),
		status_rekam_id INTEGER REFERENCES status_rekam(status_rekam_id),
		alamat_sekarang TEXT,
		status_dasar_id INTEGER REFERENCES status_dasar(status_dasar_id),
		suku_id INTEGER REFERENCES suku(suku_id),
		tag_card TEXT,
		id_asuransi_id INTEGER REFERENCES id_asuransi(id_asuransi_id),
		no_asuransi TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Create update timestamp function
	-- Fungsi trigger update timestamp
	CREATE OR REPLACE FUNCTION update_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Trigger untuk auto-update kolom updated_at
	CREATE TRIGGER update_timestamp
		BEFORE UPDATE ON data_entry_sessions
		FOR EACH ROW
		EXECUTE FUNCTION update_timestamp();
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Populate lookup tables
	if err := populateLookupTables(db); err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to populate lookup tables: %v", err)
	}

	return db, err
}

func populateLookupTables(db *sqlx.DB) error {
	if err := insertDataFromJSON(db, "json/8_sex.json", "sex", "sex_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/11_agama.json", "agama", "agama_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/12_pendidikan_kk.json", "pendidikan_kk", "pendidikan_kk_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/13_pendidikan_sedang.json", "pendidikan_sedang", "pendidikan_sedang_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/14_pekerjaan.json", "pekerjaan", "pekerjaan_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/15_status_kawin.json", "status_kawin", "status_kawin_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/16_kk_level.json", "kk_level", "kk_level_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/17_warganegara.json", "warganegara", "warganegara_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/22_golongan_darah.json", "golongan_darah", "golongan_darah_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/31_cacat.json", "cacat", "cacat_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/32_cara_kb.json", "cara_kb", "cara_kb_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/33_hamil.json", "hamil", "hamil_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/34_ktp_el.json", "ktp_el", "ktp_el_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/35_status_rekam.json", "status_rekam", "status_rekam_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/37_status_dasar.json", "status_dasar", "status_dasar_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/38_suku.json", "suku", "suku_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/40_asuransi.json", "id_asuransi", "id_asuransi_id"); err != nil {
		return err
	}
	return nil
}

func insertDataFromJSON(db *sqlx.DB, filePath, tableName, idColumn string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to read %s: %v", filePath, err)
	}

	var jsonData map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("[ERROR] Failed to parse %s: %v", filePath, err)
	}

	var records []map[string]interface{}
	for _, v := range jsonData {
		records = v
		break
	}

	for _, record := range records {
		query := fmt.Sprintf(
			"INSERT INTO %s (%s, nama) VALUES ($1, $2) ON CONFLICT (%s) DO UPDATE SET nama = EXCLUDED.nama",
			tableName, idColumn, idColumn)
		if _, err := db.Exec(query, record["id"], record["nama"]); err != nil {
			return fmt.Errorf("failed to insert into %s: %v", tableName, err)
		}
	}

	return nil
}
