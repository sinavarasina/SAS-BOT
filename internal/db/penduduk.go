package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

type DataPenduduk struct {
	JID                  string         `db:"jid"`
	NIK                  sql.NullString `db:"nik"`
	NoKK                 sql.NullString `db:"no_kk"`
	Nama                 sql.NullString `db:"nama"`
	Dusun                sql.NullString `db:"dusun"`
	RT                   sql.NullString `db:"rt"`
	SexID                sql.NullInt64  `db:"sex_id"`
	TempatLahir          sql.NullString `db:"tempat_lahir"`
	TanggalLahir         sql.NullTime   `db:"tanggal_lahir"`
	AgamaID              sql.NullInt64  `db:"agama_id"`
	PendidikanKkID       sql.NullInt64  `db:"pendidikan_kk_id"`
	PendidikanSedangID   sql.NullInt64  `db:"pendidikan_sedang_id"`
	PekerjaanID          sql.NullInt64  `db:"pekerjaan_id"`
	StatusKawinID        sql.NullInt64  `db:"status_kawin_id"`
	KkLevelID            sql.NullInt64  `db:"kk_level_id"`
	WarganegaraID        sql.NullInt64  `db:"warganegara_id"`
	NamaAyah             sql.NullString `db:"nama_ayah"`
	NamaIbu              sql.NullString `db:"nama_ibu"`
	StatusDasarID        sql.NullInt64  `db:"status_dasar_id"`
	SukuID               sql.NullInt64  `db:"suku_id"`
	NikAyah              sql.NullString `db:"nik_ayah"`
	NikIbu               sql.NullString `db:"nik_ibu"`
	GolonganDarahID      sql.NullInt64  `db:"golongan_darah_id"`
	AktaLahir            sql.NullString `db:"akta_lahir"`
	DokumenPassport      sql.NullString `db:"dokumen_passport"`
	TanggalAkhirPassport sql.NullTime   `db:"tanggal_akhir_passport"`
	DokumenKitas         sql.NullString `db:"dokumen_kitas"`
	AktaPerkawinan       sql.NullString `db:"akta_perkawinan"`
	TanggalPerkawinan    sql.NullTime   `db:"tanggal_perkawinan"`
	AktaPerceraian       sql.NullString `db:"akta_perceraian"`
	TanggalPerceraian    sql.NullTime   `db:"tanggal_perceraian"`
	CacatID              sql.NullInt64  `db:"cacat_id"`
	CaraKbID             sql.NullInt64  `db:"cara_kb_id"`
	HamilID              sql.NullInt64  `db:"hamil_id"`
	KtpElID              sql.NullInt64  `db:"ktp_el_id"`
	StatusRekamID        sql.NullInt64  `db:"status_rekam_id"`
	AlamatSekarang       sql.NullString `db:"alamat_sekarang"`
	TagCard              sql.NullString `db:"tag_card"`
	IDAsuransiID         sql.NullInt64  `db:"id_asuransi_id"`
	NoAsuransi           sql.NullString `db:"no_asuransi"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
}

// CheckNIKExistsInPenduduk adalah pengecekan NIK yang CEPAT.
func CheckNIKExistsInPenduduk(dbConn *sqlx.DB, nik string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM data_penduduk WHERE nik = $1)"
	err := dbConn.Get(&exists, query, nik)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

// GetDataPendudukByNIK mengambil data permanen berdasarkan NIK.
func GetDataPendudukByNIK(dbConn *sqlx.DB, nik string) (*DataPenduduk, error) {
	var data DataPenduduk
	query := "SELECT * FROM data_penduduk WHERE nik = $1"
	err := dbConn.Get(&data, query, nik)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteDataPendudukByNIK menghapus data permanen berdasarkan NIK.
func DeleteDataPendudukByNIK(dbConn *sqlx.DB, nik string) error {
	query := "DELETE FROM data_penduduk WHERE nik = $1"
	_, err := dbConn.Exec(query, nik)
	return err
}

// SaveDataPenduduk menyimpan (INSERT atau UPDATE) data dari sesi ke tabel permanen.
func SaveDataPenduduk(dbConn *sqlx.DB, s DataEntrySession) error {
	query := `
	INSERT INTO data_penduduk (
		jid, nik, no_kk, nama, dusun, rt, sex_id, tempat_lahir, tanggal_lahir, 
		agama_id, pendidikan_kk_id, pendidikan_sedang_id, pekerjaan_id, status_kawin_id, 
		kk_level_id, warganegara_id, nama_ayah, nama_ibu, status_dasar_id, suku_id,
		nik_ayah, nik_ibu, golongan_darah_id, akta_lahir, dokumen_passport, tanggal_akhir_passport, 
		dokumen_kitas, akta_perkawinan, tanggal_perkawinan, akta_perceraian, tanggal_perceraian, 
		cacat_id, cara_kb_id, hamil_id, ktp_el_id, status_rekam_id, alamat_sekarang, 
		tag_card, id_asuransi_id, no_asuransi, updated_at
	) VALUES (
		:jid, :nik, :no_kk, :nama, :dusun, :rt, :sex_id, :tempat_lahir, :tanggal_lahir, 
		:agama_id, :pendidikan_kk_id, :pendidikan_sedang_id, :pekerjaan_id, :status_kawin_id, 
		:kk_level_id, :warganegara_id, :nama_ayah, :nama_ibu, :status_dasar_id, :suku_id,
		:nik_ayah, :nik_ibu, :golongan_darah_id, :akta_lahir, :dokumen_passport, :tanggal_akhir_passport, 
		:dokumen_kitas, :akta_perkawinan, :tanggal_perkawinan, :akta_perceraian, :tanggal_perceraian, 
		:cacat_id, :cara_kb_id, :hamil_id, :ktp_el_id, :status_rekam_id, :alamat_sekarang, 
		:tag_card, :id_asuransi_id, :no_asuransi, NOW()
	)
	ON CONFLICT (nik) DO UPDATE SET
		jid = EXCLUDED.jid, no_kk = EXCLUDED.no_kk, nama = EXCLUDED.nama, dusun = EXCLUDED.dusun, 
		rt = EXCLUDED.rt, sex_id = EXCLUDED.sex_id, tempat_lahir = EXCLUDED.tempat_lahir, 
		tanggal_lahir = EXCLUDED.tanggal_lahir, agama_id = EXCLUDED.agama_id, 
		pendidikan_kk_id = EXCLUDED.pendidikan_kk_id, pendidikan_sedang_id = EXCLUDED.pendidikan_sedang_id, 
		pekerjaan_id = EXCLUDED.pekerjaan_id, status_kawin_id = EXCLUDED.status_kawin_id, 
		kk_level_id = EXCLUDED.kk_level_id, warganegara_id = EXCLUDED.warganegara_id, 
		nama_ayah = EXCLUDED.nama_ayah, nama_ibu = EXCLUDED.nama_ibu, 
		status_dasar_id = EXCLUDED.status_dasar_id, suku_id = EXCLUDED.suku_id,
		nik_ayah = EXCLUDED.nik_ayah, nik_ibu = EXCLUDED.nik_ibu, 
		golongan_darah_id = EXCLUDED.golongan_darah_id, akta_lahir = EXCLUDED.akta_lahir, 
		dokumen_passport = EXCLUDED.dokumen_passport, tanggal_akhir_passport = EXCLUDED.tanggal_akhir_passport, 
		dokumen_kitas = EXCLUDED.dokumen_kitas, akta_perkawinan = EXCLUDED.akta_perkawinan, 
		tanggal_perkawinan = EXCLUDED.tanggal_perkawinan, akta_perceraian = EXCLUDED.akta_perceraian, 
		tanggal_perceraian = EXCLUDED.tanggal_perceraian, cacat_id = EXCLUDED.cacat_id, 
		cara_kb_id = EXCLUDED.cara_kb_id, hamil_id = EXCLUDED.hamil_id, ktp_el_id = EXCLUDED.ktp_el_id, 
		status_rekam_id = EXCLUDED.status_rekam_id, alamat_sekarang = EXCLUDED.alamat_sekarang, 
		tag_card = EXCLUDED.tag_card, id_asuransi_id = EXCLUDED.id_asuransi_id, 
		no_asuransi = EXCLUDED.no_asuransi, updated_at = NOW()
	`
	_, err := dbConn.NamedExec(query, s)
	if err != nil {
		log.Printf("Gagal SaveDataPenduduk: %v", err)
	}
	return err
}

// LoadSessionFromPenduduk menyalin data dari tabel permanen ke sesi sementara untuk diedit.
func LoadSessionFromPenduduk(dbConn *sqlx.DB, jid string, data DataPenduduk) error {
	query := `
	INSERT INTO data_entry_sessions (
		jid, current_step, awaiting_answer,
		nik, no_kk, nama, dusun, rt, sex_id, tempat_lahir, tanggal_lahir, 
		agama_id, pendidikan_kk_id, pendidikan_sedang_id, pekerjaan_id, status_kawin_id, 
		kk_level_id, warganegara_id, nama_ayah, nama_ibu, status_dasar_id, suku_id,
		nik_ayah, nik_ibu, golongan_darah_id, akta_lahir, dokumen_passport, tanggal_akhir_passport, 
		dokumen_kitas, akta_perkawinan, tanggal_perkawinan, akta_perceraian, tanggal_perceraian, 
		cacat_id, cara_kb_id, hamil_id, ktp_el_id, status_rekam_id, alamat_sekarang, 
		tag_card, id_asuransi_id, no_asuransi,
		created_at, updated_at, sheet_row_num, edit_field
	) VALUES (
		$1, $2, $3,
		$4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22,
		$23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42,
		NOW(), NOW(), NULL, NULL
	)
	ON CONFLICT (jid) DO UPDATE SET
		current_step = EXCLUDED.current_step, awaiting_answer = EXCLUDED.awaiting_answer,
		nik = EXCLUDED.nik, no_kk = EXCLUDED.no_kk, nama = EXCLUDED.nama, dusun = EXCLUDED.dusun, 
		rt = EXCLUDED.rt, sex_id = EXCLUDED.sex_id, tempat_lahir = EXCLUDED.tempat_lahir, 
		tanggal_lahir = EXCLUDED.tanggal_lahir, agama_id = EXCLUDED.agama_id, 
		pendidikan_kk_id = EXCLUDED.pendidikan_kk_id, pendidikan_sedang_id = EXCLUDED.pendidikan_sedang_id, 
		pekerjaan_id = EXCLUDED.pekerjaan_id, status_kawin_id = EXCLUDED.status_kawin_id, 
		kk_level_id = EXCLUDED.kk_level_id, warganegara_id = EXCLUDED.warganegara_id, 
		nama_ayah = EXCLUDED.nama_ayah, nama_ibu = EXCLUDED.nama_ibu, 
		status_dasar_id = EXCLUDED.status_dasar_id, suku_id = EXCLUDED.suku_id,
		nik_ayah = EXCLUDED.nik_ayah, nik_ibu = EXCLUDED.nik_ibu, 
		golongan_darah_id = EXCLUDED.golongan_darah_id, akta_lahir = EXCLUDED.akta_lahir, 
		dokumen_passport = EXCLUDED.dokumen_passport, tanggal_akhir_passport = EXCLUDED.tanggal_akhir_passport, 
		dokumen_kitas = EXCLUDED.dokumen_kitas, akta_perkawinan = EXCLUDED.akta_perkawinan, 
		tanggal_perkawinan = EXCLUDED.tanggal_perkawinan, akta_perceraian = EXCLUDED.akta_perceraian, 
		tanggal_perceraian = EXCLUDED.tanggal_perceraian, cacat_id = EXCLUDED.cacat_id, 
		cara_kb_id = EXCLUDED.cara_kb_id, hamil_id = EXCLUDED.hamil_id, ktp_el_id = EXCLUDED.ktp_el_id, 
		status_rekam_id = EXCLUDED.status_rekam_id, alamat_sekarang = EXCLUDED.alamat_sekarang, 
		tag_card = EXCLUDED.tag_card, id_asuransi_id = EXCLUDED.id_asuransi_id, 
		no_asuransi = EXCLUDED.no_asuransi, updated_at = NOW(),
		created_at = EXCLUDED.created_at, sheet_row_num = NULL, edit_field = NULL
	`
	_, err := dbConn.Exec(query,
		jid, 42, true,
		data.NIK, data.NoKK, data.Nama, data.Dusun, data.RT, data.SexID,
		data.TempatLahir, data.TanggalLahir, data.AgamaID, data.PendidikanKkID,
		data.PendidikanSedangID, data.PekerjaanID, data.StatusKawinID, data.KkLevelID,
		data.WarganegaraID, data.NamaAyah, data.NamaIbu, data.StatusDasarID, data.SukuID,
		data.NikAyah, data.NikIbu, data.GolonganDarahID, data.AktaLahir, data.DokumenPassport,
		data.TanggalAkhirPassport, data.DokumenKitas, data.AktaPerkawinan,
		data.TanggalPerkawinan, data.AktaPerceraian, data.TanggalPerceraian, data.CacatID,
		data.CaraKbID, data.HamilID, data.KtpElID, data.StatusRekamID, data.AlamatSekarang,
		data.TagCard, data.IDAsuransiID, data.NoAsuransi,
	)
	return err
}
