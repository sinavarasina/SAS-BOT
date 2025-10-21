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
	CREATE TABLE IF NOT EXISTS data_penduduk (
    -- Kunci primer untuk tabel ini
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    
    -- ID untuk pencatat/operator data entry (sesuai permintaan Anda "setiap user_id")
    user_id INT, 
    
    -- Kolom dari 1 sampai 7
    alamat TEXT,
    dusun VARCHAR(100),
    rw VARCHAR(5),
    rt VARCHAR(5),
    nama VARCHAR(255) NOT NULL,
    no_kk VARCHAR(16),
    nik VARCHAR(16) NOT NULL UNIQUE,
    
    -- Kolom 8 (Foreign Key ke tabel 'sex')
    sex_id INT,
    
    -- Kolom 9 & 10
    tempat_lahir VARCHAR(100),
    tanggal_lahir DATE,
    
    -- Kolom 11 - 17 (Foreign Keys)
    agama_id INT,
    pendidikan_kk_id INT,
    pendidikan_sedang_id INT,
    pekerjaan_id INT,
    status_kawin_id INT,
    kk_level_id INT,
    warganegara_id INT,
    
    -- Kolom 18 - 21
    nik_ayah VARCHAR(16),
    nama_ayah VARCHAR(255),
    nik_ibu VARCHAR(16),
    nama_ibu VARCHAR(255),
    
    -- Kolom 22 (Foreign Key)
    golongan_darah_id INT,
    
    -- Kolom 23 - 30
    akta_lahir VARCHAR(100),
    dokumen_passport VARCHAR(50),
    tanggal_akhir_passport DATE,
    dokumen_kitas VARCHAR(50),
    akta_perkawinan VARCHAR(100),
    tanggal_perkawinan DATE,
    akta_perceraian VARCHAR(100), -- Mengganti 'pencarian' menjadi 'perceraian'
    tanggal_perceraian DATE,
    
    -- Kolom 31 - 35 (Foreign Keys)
    cacat_id INT,
    cara_kb_id INT,
    hamil_id INT,
    ktp_el_id INT, -- Mengganti 'ktp_elektronik' menjadi 'ktp_el_id'
    status_rekam_id INT,
    
    -- Kolom 36
    alamat_sekarang TEXT,
    
    -- Kolom 37 - 40 (Foreign Keys)
    status_dasar_id INT,
    suku_id INT,
    tag_card VARCHAR(100),
    id_asuransi_id INT, -- Mengganti 'asuransi' menjadi 'id_asuransi_id'

    -- Kolom audit (opsional tapi sangat disarankan)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- --- DEKLARASI FOREIGN KEY ---
    -- (Asumsi ada tabel 'users' untuk data entry)
    -- FOREIGN KEY (user_id) REFERENCES users(id), 
    
    FOREIGN KEY (sex_id) REFERENCES sex(sex_id),
    FOREIGN KEY (agama_id) REFERENCES agama(agama_id),
    FOREIGN KEY (pendidikan_kk_id) REFERENCES pendidikan_kk(pendidikan_kk_id),
    FOREIGN KEY (pendidikan_sedang_id) REFERENCES pendidikan_sedang(pendidikan_sedang_id),
    FOREIGN KEY (pekerjaan_id) REFERENCES pekerjaan(pekerjaan_id),
    FOREIGN KEY (status_kawin_id) REFERENCES status_kawin(status_kawin_id),
    FOREIGN KEY (kk_level_id) REFERENCES kk_level(kk_level_id),
    FOREIGN KEY (warganegara_id) REFERENCES warganegara(warganegara_id),
    FOREIGN KEY (golongan_darah_id) REFERENCES golongan_darah(golongan_darah_id),
    FOREIGN KEY (cacat_id) REFERENCES cacat(cacat_id),
    FOREIGN KEY (cara_kb_id) REFERENCES cara_kb(cara_kb_id),
    FOREIGN KEY (hamil_id) REFERENCES hamil(hamil_id),
    FOREIGN KEY (ktp_el_id) REFERENCES ktp_el(ktp_el_id),
    FOREIGN KEY (status_rekam_id) REFERENCES status_rekam(status_rekam_id),
    FOREIGN KEY (status_dasar_id) REFERENCES status_dasar(status_dasar_id),
    FOREIGN KEY (suku_id) REFERENCES SUKU(suku_id),
    FOREIGN KEY (id_asuransi_id) REFERENCES id_asuransi(id_asuransi_id)
);
	`
	_, err = db.Exec(schema)
	return db, err
}
