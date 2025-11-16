package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	schema := `
	-- Fungsi & Trigger
	CREATE OR REPLACE FUNCTION update_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS update_timestamp_sessions ON data_entry_sessions;
	DROP TRIGGER IF EXISTS update_timestamp_penduduk ON data_penduduk;

	-- Tabel-tabel dasar
	CREATE TABLE IF NOT EXISTS users (
		jid TEXT PRIMARY KEY,
		number TEXT,
		username TEXT,
		previlege TEXT
	);
	CREATE TABLE IF NOT EXISTS pengaduan (
		id SERIAL PRIMARY KEY,
		user_jid TEXT REFERENCES users(jid),
		deskripsi TEXT,
		pict_path TEXT,
		sent_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Tabel Lookup (17 tabel)
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
		current_flow TEXT,
		sheet_row_num INTEGER,
		dusun TEXT,
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
		nama_ayah TEXT,
		nama_ibu TEXT,
		status_dasar_id INTEGER REFERENCES status_dasar(status_dasar_id),
		suku_id INTEGER REFERENCES suku(suku_id),
		nik_ayah TEXT,
		nik_ibu TEXT,
		
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
		tag_card TEXT,
		id_asuransi_id INTEGER REFERENCES id_asuransi(id_asuransi_id),
		no_asuransi TEXT,
		
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		
		surat_valid_nik TEXT,
		surat_fields_pending TEXT,
		surat_field_now TEXT,
		surat_data_map TEXT,
		surat_temp_answer TEXT,
		
		edit_field TEXT
	);

	CREATE TABLE IF NOT EXISTS data_penduduk (
		jid TEXT,
		nik TEXT PRIMARY KEY,
		no_kk TEXT,
		nama TEXT,
		dusun TEXT,
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
		nama_ayah TEXT,
		nama_ibu TEXT,
		status_dasar_id INTEGER REFERENCES status_dasar(status_dasar_id),
		suku_id INTEGER REFERENCES suku(suku_id),
		nik_ayah TEXT,
		nik_ibu TEXT,
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
		tag_card TEXT,
		id_asuransi_id INTEGER REFERENCES id_asuransi(id_asuransi_id),
		no_asuransi TEXT,
		google_token TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Trigger
	CREATE TRIGGER update_timestamp_sessions
		BEFORE UPDATE ON data_entry_sessions
		FOR EACH ROW
		EXECUTE FUNCTION update_timestamp();
		
	CREATE TRIGGER update_timestamp_penduduk
		BEFORE UPDATE ON data_penduduk
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

	if err := EnsureFlowColumn(db); err != nil { return nil, err }
	if err := EnsureEditFieldColumn(db); err != nil { return nil, err }
	if err := EnsureSheetRowNumColumn(db); err != nil { return nil, err }
	if err := EnsureSuratSessionColumns(db); err != nil { return nil, err }

	if err := DropColumnIfExists(db, "data_entry_sessions", "current_menu"); err != nil { return nil, err }

	return db, err
}

func populateLookupTables(db *sqlx.DB) error {
	if err := insertDataFromJSON(db, "json/6_sex.json", "sex", "sex_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/9_agama.json", "agama", "agama_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/10_pendidikan_kk.json", "pendidikan_kk", "pendidikan_kk_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/11_pendidikan_sedang.json", "pendidikan_sedang", "pendidikan_sedang_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/12_pekerjaan.json", "pekerjaan", "pekerjaan_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/13_status_kawin.json", "status_kawin", "status_kawin_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/14_kk_level.json", "kk_level", "kk_level_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/15_warganegara.json", "warganegara", "warganegara_id"); err != nil {
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
	if err := insertDataFromJSON(db, "json/18_status_dasar.json", "status_dasar", "status_dasar_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/19_suku.json", "suku", "suku_id"); err != nil {
		return err
	}
	if err := insertDataFromJSON(db, "json/38_asuransi.json", "id_asuransi", "id_asuransi_id"); err != nil {
		return err
	}
	return nil
}

func insertDataFromJSON(db *sqlx.DB, filePath, tableName, idColumn string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("[ERROR] Gagal membaca %s: %v", filePath, err)
	}

	var jsonData map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("[ERROR] Gagal parse %s: %v", filePath, err)
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
			return fmt.Errorf("gagal insert ke %s: %v", tableName, err)
		}
	}

	return nil
}

func DropColumnIfExists(dbConn *sqlx.DB, tableName, columnName string) error {
	var exists bool
	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = '%s' 
			AND column_name = '%s'
		);
	`, tableName, columnName)
	err := dbConn.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check %s column: %v", columnName, err)
	}

	if exists {
		log.Printf("[DEBUG] Column '%s' ditemukan di table '%s', menghapusnya...", columnName, tableName)
		_, err = dbConn.Exec(fmt.Sprintf(`
			ALTER TABLE %s 
			DROP COLUMN %s;
		`, tableName, columnName))
		if err != nil {
			return fmt.Errorf("failed to drop %s column: %v", columnName, err)
		}
		log.Printf("[DEBUG] Berhasil menghapus %s column dari %s table", columnName, tableName)
	}
	return nil
}
