package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type DataEntrySession struct {
	JID                  string         `db:"jid"`
	CurrentStep          int            `db:"current_step"`
	Alamat               sql.NullString `db:"alamat"`
	Dusun                sql.NullString `db:"dusun"`
	RW                   sql.NullString `db:"rw"`
	RT                   sql.NullString `db:"rt"`
	Nama                 sql.NullString `db:"nama"`
	NoKK                 sql.NullString `db:"no_kk"`
	NIK                  sql.NullString `db:"nik"`
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
	NikAyah              sql.NullString `db:"nik_ayah"`
	NamaAyah             sql.NullString `db:"nama_ayah"`
	NikIbu               sql.NullString `db:"nik_ibu"`
	NamaIbu              sql.NullString `db:"nama_ibu"`
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
	StatusDasarID        sql.NullInt64  `db:"status_dasar_id"`
	SukuID               sql.NullInt64  `db:"suku_id"`
	TagCard              sql.NullString `db:"tag_card"`
	IDAsuransiID         sql.NullInt64  `db:"id_asuransi_id"`
	NoAsuransi           sql.NullString `db:"no_asuransi"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	AwaitingAnswer        bool           `db:"awaiting_answer"`
}

// GetOrCreateDataEntrySession retrieves an existing session or creates a new one.
func GetOrCreateDataEntrySession(dbConn *sqlx.DB, jid string) (*DataEntrySession, error) {
	var session DataEntrySession
	err := dbConn.Get(&session, "SELECT * FROM data_entry_sessions WHERE jid = $1", jid)
	if err == sql.ErrNoRows {
		log.Printf("[DEBUG] Creating new session for jid: %s", jid)
		newSession := DataEntrySession{JID: jid, CurrentStep: 1, AwaitingAnswer: false}
		_, err := dbConn.NamedExec(`
			INSERT INTO data_entry_sessions (jid, current_step, awaiting_answer) 
			VALUES (:jid, :current_step, :awaiting_answer)`, newSession)
		if err != nil {
			log.Printf("[ERROR] Failed to create session: %v", err)
			return nil, err
		}
		log.Printf("[DEBUG] New session created successfully")
		return &newSession, nil
	} else if err != nil {
		log.Printf("[ERROR] Error fetching session: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Existing session found - Step: %d, Awaiting: %v", session.CurrentStep, session.AwaitingAnswer)
	return &session, nil
}

func StartNewSession(dbConn *sqlx.DB, jid string) error {
	log.Printf("[DEBUG] Starting new session for jid: %s", jid)

	// Clear all data and reset the session
	_, err := dbConn.Exec(`
        UPDATE data_entry_sessions 
        SET current_step = 1,
            awaiting_answer = true,
            alamat = NULL,
            dusun = NULL,
            rw = NULL,
            rt = NULL,
            nama = NULL,
            no_kk = NULL,
            nik = NULL,
            sex_id = NULL,
            tempat_lahir = NULL,
            tanggal_lahir = NULL,
            agama_id = NULL,
            pendidikan_kk_id = NULL,
            pendidikan_sedang_id = NULL,
            pekerjaan_id = NULL,
            status_kawin_id = NULL,
            kk_level_id = NULL,
            warganegara_id = NULL,
            nik_ayah = NULL,
            nama_ayah = NULL,
            nik_ibu = NULL,
            nama_ibu = NULL,
            golongan_darah_id = NULL,
            akta_lahir = NULL,
            dokumen_passport = NULL,
            tanggal_akhir_passport = NULL,
            dokumen_kitas = NULL,
            akta_perkawinan = NULL,
            tanggal_perkawinan = NULL,
            akta_perceraian = NULL,
            tanggal_perceraian = NULL,
            cacat_id = NULL,
            cara_kb_id = NULL,
            hamil_id = NULL,
            ktp_el_id = NULL,
            status_rekam_id = NULL,
            alamat_sekarang = NULL,
            status_dasar_id = NULL,
            suku_id = NULL,
            tag_card = NULL,
            id_asuransi_id = NULL,
            no_asuransi = NULL,
            updated_at = NOW()
        WHERE jid = $1`, jid)

	if err != nil {
		log.Printf("[ERROR] Failed to reset session: %v", err)
		return err
	}

	log.Printf("[DEBUG] Session started/reset successfully")
	return nil
}

// UpdateDataEntrySession updates a specific field in the session and increments the step.
func UpdateDataEntrySession(dbConn *sqlx.DB, jid string, field string, value interface{}) error {
	query := `UPDATE data_entry_sessions 
              SET ` + field + ` = $1, 
                  current_step = current_step + 1,
                  awaiting_answer = true
              WHERE jid = $2`
	_, err := dbConn.Exec(query, value, jid)
	return err
}

// DeleteDataEntrySession removes a session.
func DeleteDataEntrySession(dbConn *sqlx.DB, jid string) error {
	log.Printf("[DEBUG] Deleting session for jid: %s", jid)
	result, err := dbConn.Exec("DELETE FROM data_entry_sessions WHERE jid = $1", jid)
	if err != nil {
		log.Printf("[ERROR] Failed to delete session: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Failed to get rows affected: %v", err)
		return err
	}

	log.Printf("[DEBUG] Deleted %d session rows", rows)
	return nil
}
