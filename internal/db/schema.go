package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		jid TEXT PRIMARY KEY,
		number TEXT,
		username TEXT,
		previlege TEXT
	);

	CREATE TABLE IF NOT EXISTS user_nik (
		jid TEXT PRIMARY KEY,
		nik TEXT
	);

	-- Lookup Tables from format_sas.json
	CREATE TABLE IF NOT EXISTS sex ( sex_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS agama ( agama_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS pendidikan_kk ( pendidikan_kk_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS pendidikan_sedang ( pendidikan_sedang_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS pekerjaan ( pekerjaan_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS status_kawin ( status_kawin_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS kk_level ( kk_level_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS warganegara ( warganegara_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS golongan_darah ( golongan_darah_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS cacat ( cacat_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS cara_kb ( cara_kb_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS hamil ( hamil_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS ktp_el ( ktp_el_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS status_rekam ( status_rekam_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS status_dasar ( status_dasar_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS suku ( suku_id SERIAL PRIMARY KEY, nama TEXT );
	CREATE TABLE IF NOT EXISTS id_asuransi ( id_asuransi_id SERIAL PRIMARY KEY, nama TEXT );

	CREATE TABLE IF NOT EXISTS data_entry_sessions (
		jid TEXT PRIMARY KEY,
		current_step INT DEFAULT 1,
		awaiting_answer BOOLEAN DEFAULT false,
		alamat TEXT,
		dusun VARCHAR(100),
		rw VARCHAR(5),
		rt VARCHAR(5),
		nama VARCHAR(255),
		no_kk VARCHAR(16),
		nik VARCHAR(16),
		sex_id INT,
		tempat_lahir VARCHAR(100),
		tanggal_lahir DATE,
		agama_id INT,
		pendidikan_kk_id INT,
		pendidikan_sedang_id INT,
		pekerjaan_id INT,
		status_kawin_id INT,
		kk_level_id INT,
		warganegara_id INT,
		nik_ayah VARCHAR(16),
		nama_ayah VARCHAR(255),
		nik_ibu VARCHAR(16),
		nama_ibu VARCHAR(255),
		golongan_darah_id INT,
		akta_lahir VARCHAR(100),
		dokumen_passport VARCHAR(50),
		tanggal_akhir_passport DATE,
		dokumen_kitas VARCHAR(50),
		akta_perkawinan VARCHAR(100),
		tanggal_perkawinan DATE,
		akta_perceraian VARCHAR(100),
		tanggal_perceraian DATE,
		cacat_id INT,
		cara_kb_id INT,
		hamil_id INT,
		ktp_el_id INT,
		status_rekam_id INT,
		alamat_sekarang TEXT,
		status_dasar_id INT,
		suku_id INT,
		tag_card VARCHAR(100),
		id_asuransi_id INT,
		no_asuransi TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE OR REPLACE FUNCTION update_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_timestamp ON data_entry_sessions;
	CREATE TRIGGER update_timestamp
		BEFORE UPDATE ON data_entry_sessions
		FOR EACH ROW
		EXECUTE PROCEDURE update_timestamp();
	`
	_, err = db.Exec(schema)
	return db, err
}
