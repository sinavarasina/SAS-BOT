package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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
	AwaitingAnswer       bool           `db:"awaiting_answer"`
	EditField            sql.NullString `db:"edit_field"`

	// Change lookup table name fields to use sql.NullString
	SexNama              sql.NullString `db:"sex_nama"`
	AgamaNama            sql.NullString `db:"agama_nama"`
	PendidikanKKNama     sql.NullString `db:"pendidikan_kk_nama"`
	PendidikanSedangNama sql.NullString `db:"pendidikan_sedang_nama"`
	PekerjaanNama        sql.NullString `db:"pekerjaan_nama"`
	StatusKawinNama      sql.NullString `db:"status_kawin_nama"`
	KKLevelNama          sql.NullString `db:"kk_level_nama"`
	WarganegaraNama      sql.NullString `db:"warganegara_nama"`
	GolonganDarahNama    sql.NullString `db:"golongan_darah_nama"`
	CacatNama            sql.NullString `db:"cacat_nama"`
	CaraKBNama           sql.NullString `db:"cara_kb_nama"`
	HamilNama            sql.NullString `db:"hamil_nama"`
	KTPElNama            sql.NullString `db:"ktp_el_nama"`
	StatusRekamNama      sql.NullString `db:"status_rekam_nama"`
	StatusDasarNama      sql.NullString `db:"status_dasar_nama"`
	SukuNama             sql.NullString `db:"suku_nama"`
	AsuransiNama         sql.NullString `db:"asuransi_nama"`
}

// GetOrCreateDataEntrySession retrieves an existing session or creates a new one.
func GetOrCreateDataEntrySession(dbConn *sqlx.DB, jid string) (*DataEntrySession, error) {
	// Ensure edit_field column exists
	if err := EnsureEditFieldColumn(dbConn); err != nil {
		log.Printf("[ERROR] Failed to ensure edit_field column: %v", err)
		return nil, err
	}

	var session DataEntrySession
	err := dbConn.Get(&session, "SELECT * FROM data_entry_sessions WHERE jid = $1", jid)
	if err == sql.ErrNoRows {
		log.Printf("[DEBUG] Creating new session for jid: %s", jid)
		newSession := DataEntrySession{
			JID:            jid,
			CurrentStep:    1,
			AwaitingAnswer: false,
			EditField:      sql.NullString{String: "", Valid: false},
		}
		_, err := dbConn.NamedExec(`
			INSERT INTO data_entry_sessions (jid, current_step, awaiting_answer, edit_field) 
			VALUES (:jid, :current_step, :awaiting_answer, :edit_field)`, newSession)
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

	// Updated query to set current_step = 1 and awaiting_answer = true
	_, err := dbConn.Exec(`
        INSERT INTO data_entry_sessions (jid, current_step, awaiting_answer, created_at, updated_at)
        VALUES ($1, 1, true, NOW(), NOW())
        ON CONFLICT (jid) DO UPDATE 
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
            edit_field = NULL,
            updated_at = NOW()`, jid)

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

func GetFullSessionData(dbConn *sqlx.DB, jid string) (*DataEntrySession, error) {
	log.Printf("[DEBUG] Getting full session data for jid: %s", jid)
	var session DataEntrySession
	// Query ini sama dengan query di GetFormattedSessionData
	query := `
        SELECT s.*, 
            sex.nama as sex_nama,
            ag.nama as agama_nama,
            pk.nama as pendidikan_kk_nama,
            ps.nama as pendidikan_sedang_nama,
            p.nama as pekerjaan_nama,
            sk.nama as status_kawin_nama,
            kk.nama as kk_level_nama,
            w.nama as warganegara_nama,
            gd.nama as golongan_darah_nama,
            c.nama as cacat_nama,
            kb.nama as cara_kb_nama,
            h.nama as hamil_nama,
            ke.nama as ktp_el_nama,
            sr.nama as status_rekam_nama,
            sd.nama as status_dasar_nama,
            su.nama as suku_nama,
            a.nama as asuransi_nama
        FROM data_entry_sessions s
        LEFT JOIN sex ON s.sex_id = sex.sex_id
        LEFT JOIN agama ag ON s.agama_id = ag.agama_id
        LEFT JOIN pendidikan_kk pk ON s.pendidikan_kk_id = pk.pendidikan_kk_id
        LEFT JOIN pendidikan_sedang ps ON s.pendidikan_sedang_id = ps.pendidikan_sedang_id
        LEFT JOIN pekerjaan p ON s.pekerjaan_id = p.pekerjaan_id
        LEFT JOIN status_kawin sk ON s.status_kawin_id = sk.status_kawin_id
        LEFT JOIN kk_level kk ON s.kk_level_id = kk.kk_level_id
        LEFT JOIN warganegara w ON s.warganegara_id = w.warganegara_id
        LEFT JOIN golongan_darah gd ON s.golongan_darah_id = gd.golongan_darah_id
        LEFT JOIN cacat c ON s.cacat_id = c.cacat_id
        LEFT JOIN cara_kb kb ON s.cara_kb_id = kb.cara_kb_id
        LEFT JOIN hamil h ON s.hamil_id = h.hamil_id
        LEFT JOIN ktp_el ke ON s.ktp_el_id = ke.ktp_el_id
        LEFT JOIN status_rekam sr ON s.status_rekam_id = sr.status_rekam_id
        LEFT JOIN status_dasar sd ON s.status_dasar_id = sd.status_dasar_id
        LEFT JOIN suku su ON s.suku_id = su.suku_id
        LEFT JOIN id_asuransi a ON s.id_asuransi_id = a.id_asuransi_id
        WHERE s.jid = $1`

	err := dbConn.Get(&session, query, jid)
	if err != nil {
		log.Printf("[ERROR] Failed to get full session data: %v", err)
		return nil, err
	}
	return &session, nil
}

func GetFormattedSessionData(dbConn *sqlx.DB, jid string) (string, error) {
	log.Printf("[DEBUG] Getting formatted data for jid: %s", jid)

	var session DataEntrySession
	query := `
        SELECT s.*, 
            sex.nama as sex_nama,
            ag.nama as agama_nama,
            pk.nama as pendidikan_kk_nama,
            ps.nama as pendidikan_sedang_nama,
            p.nama as pekerjaan_nama,
            sk.nama as status_kawin_nama,
            kk.nama as kk_level_nama,
            w.nama as warganegara_nama,
            gd.nama as golongan_darah_nama,
            c.nama as cacat_nama,
            kb.nama as cara_kb_nama,
            h.nama as hamil_nama,
            ke.nama as ktp_el_nama,
            sr.nama as status_rekam_nama,
            sd.nama as status_dasar_nama,
            su.nama as suku_nama,
            a.nama as asuransi_nama
        FROM data_entry_sessions s
        LEFT JOIN sex ON s.sex_id = sex.sex_id
        LEFT JOIN agama ag ON s.agama_id = ag.agama_id
        LEFT JOIN pendidikan_kk pk ON s.pendidikan_kk_id = pk.pendidikan_kk_id
        LEFT JOIN pendidikan_sedang ps ON s.pendidikan_sedang_id = ps.pendidikan_sedang_id
        LEFT JOIN pekerjaan p ON s.pekerjaan_id = p.pekerjaan_id
        LEFT JOIN status_kawin sk ON s.status_kawin_id = sk.status_kawin_id
        LEFT JOIN kk_level kk ON s.kk_level_id = kk.kk_level_id
        LEFT JOIN warganegara w ON s.warganegara_id = w.warganegara_id
        LEFT JOIN golongan_darah gd ON s.golongan_darah_id = gd.golongan_darah_id
        LEFT JOIN cacat c ON s.cacat_id = c.cacat_id
        LEFT JOIN cara_kb kb ON s.cara_kb_id = kb.cara_kb_id
        LEFT JOIN hamil h ON s.hamil_id = h.hamil_id
        LEFT JOIN ktp_el ke ON s.ktp_el_id = ke.ktp_el_id
        LEFT JOIN status_rekam sr ON s.status_rekam_id = sr.status_rekam_id
        LEFT JOIN status_dasar sd ON s.status_dasar_id = sd.status_dasar_id
        LEFT JOIN suku su ON s.suku_id = su.suku_id
        LEFT JOIN id_asuransi a ON s.id_asuransi_id = a.id_asuransi_id
        WHERE s.jid = $1`

	err := dbConn.Get(&session, query, jid)
	if err != nil {
		log.Printf("[ERROR] Failed to get session data: %v", err)
		return "", err
	}

	// Add debug logging for IDs and names
	log.Printf("[DEBUG] Session data - sex_id: %v, sex_nama: %v",
		session.SexID.Int64, session.SexNama.String)
	log.Printf("[DEBUG] Session data - agama_id: %v, agama_nama: %v",
		session.AgamaID.Int64, session.AgamaNama.String)
	log.Printf("[DEBUG] Session data - pendidikan_kk_id: %v, pendidikan_kk_nama: %v",
		session.PendidikanKkID.Int64, session.PendidikanKKNama.String)
	log.Printf("[DEBUG] Session data - pendidikan_sedang_id: %v, pendidikan_sedang_nama: %v",
		session.PendidikanSedangID.Int64, session.PendidikanSedangNama.String)
	log.Printf("[DEBUG] Session data - pekerjaan_id: %v, pekerjaan_nama: %v",
		session.PekerjaanID.Int64, session.PekerjaanNama.String)
	log.Printf("[DEBUG] Session data - status_kawin_id: %v, status_kawin_nama: %v",
		session.StatusKawinID.Int64, session.StatusKawinNama.String)
	log.Printf("[DEBUG] Session data - kk_level_id: %v, kk_level_nama: %v",
		session.KkLevelID.Int64, session.KKLevelNama.String)
	log.Printf("[DEBUG] Session data - warganegara_id: %v, warganegara_nama: %v",
		session.WarganegaraID.Int64, session.WarganegaraNama.String)
	log.Printf("[DEBUG] Session data - golongan_darah_id: %v, golongan_darah_nama: %v",
		session.GolonganDarahID.Int64, session.GolonganDarahNama.String)
	log.Printf("[DEBUG] Session data - cacat_id: %v, cacat_nama: %v",
		session.CacatID.Int64, session.CacatNama.String)
	log.Printf("[DEBUG] Session data - cara_kb_id: %v, cara_kb_nama: %v",
		session.CaraKbID.Int64, session.CaraKBNama.String)
	log.Printf("[DEBUG] Session data - hamil_id: %v, hamil_nama: %v",
		session.HamilID.Int64, session.HamilNama.String)
	log.Printf("[DEBUG] Session data - ktp_el_id: %v, ktp_el_nama: %v",
		session.KtpElID.Int64, session.KTPElNama.String)
	log.Printf("[DEBUG] Session data - status_rekam_id: %v, status_rekam_nama: %v",
		session.StatusRekamID.Int64, session.StatusRekamNama.String)
	log.Printf("[DEBUG] Session data - status_dasar_id: %v, status_dasar_nama: %v",
		session.StatusDasarID.Int64, session.StatusDasarNama.String)
	log.Printf("[DEBUG] Session data - suku_id: %v, suku_nama: %v",
		session.SukuID.Int64, session.SukuNama.String)
	log.Printf("[DEBUG] Session data - id_asuransi_id: %v, asuransi_nama: %v",
		session.IDAsuransiID.Int64, session.AsuransiNama.String)

	var result strings.Builder

	// Modified appendNumberedField to show names for lookup fields
	appendNumberedField := func(num int, label string, value interface{}, name sql.NullString) {
		if str, ok := value.(sql.NullInt64); ok && str.Valid {
			if name.Valid && name.String != "" {
				// For lookup fields, show the name instead of ID
				fmt.Fprintf(&result, "%d. %s: %s\n", num, label, name.String)
			} else {
				fmt.Fprintf(&result, "%d. %s: %d\n", num, label, str.Int64)
			}
		} else if str, ok := value.(sql.NullString); ok && str.Valid {
			fmt.Fprintf(&result, "%d. %s: %s\n", num, label, str.String)
		} else if str, ok := value.(sql.NullTime); ok && str.Valid {
			fmt.Fprintf(&result, "%d. %s: %s\n", num, label, FormatDate(str))
		} else {
			fmt.Fprintf(&result, "%d. %s: -\n", num, label)
		}
	}

	// Always show all fields with their numbers and IDs
	appendNumberedField(1, "Alamat", session.Alamat, sql.NullString{})
	appendNumberedField(2, "Dusun", session.Dusun, sql.NullString{})
	appendNumberedField(3, "RW", session.RW, sql.NullString{})
	appendNumberedField(4, "RT", session.RT, sql.NullString{})
	appendNumberedField(5, "Nama", session.Nama, sql.NullString{})
	appendNumberedField(6, "No KK", session.NoKK, sql.NullString{})
	appendNumberedField(7, "NIK", session.NIK, sql.NullString{})
	appendNumberedField(8, "Jenis Kelamin", session.SexID, session.SexNama)
	appendNumberedField(9, "Tempat Lahir", session.TempatLahir, sql.NullString{})
	appendNumberedField(10, "Tanggal Lahir", session.TanggalLahir, sql.NullString{})
	appendNumberedField(11, "Agama", session.AgamaID, session.AgamaNama)
	appendNumberedField(12, "Pendidikan KK", session.PendidikanKkID, session.PendidikanKKNama)
	appendNumberedField(13, "Pendidikan Sedang", session.PendidikanSedangID, session.PendidikanSedangNama)
	appendNumberedField(14, "Pekerjaan", session.PekerjaanID, session.PekerjaanNama)
	appendNumberedField(15, "Status Kawin", session.StatusKawinID, session.StatusKawinNama)
	appendNumberedField(16, "Level KK", session.KkLevelID, session.KKLevelNama)
	appendNumberedField(17, "Warganegara", session.WarganegaraID, session.WarganegaraNama)
	appendNumberedField(18, "NIK Ayah", session.NikAyah, sql.NullString{})
	appendNumberedField(19, "Nama Ayah", session.NamaAyah, sql.NullString{})
	appendNumberedField(20, "NIK Ibu", session.NikIbu, sql.NullString{})
	appendNumberedField(21, "Nama Ibu", session.NamaIbu, sql.NullString{})
	appendNumberedField(22, "Golongan Darah", session.GolonganDarahID, session.GolonganDarahNama)
	appendNumberedField(23, "No. Akta Lahir", session.AktaLahir, sql.NullString{})
	appendNumberedField(24, "No. Paspor", session.DokumenPassport, sql.NullString{})
	appendNumberedField(25, "Tgl Akhir Paspor", session.TanggalAkhirPassport, sql.NullString{})
	appendNumberedField(26, "No. KITAS", session.DokumenKitas, sql.NullString{})
	appendNumberedField(27, "No. Akta Kawin", session.AktaPerkawinan, sql.NullString{})
	appendNumberedField(28, "Tgl Perkawinan", session.TanggalPerkawinan, sql.NullString{})
	appendNumberedField(29, "No. Akta Cerai", session.AktaPerceraian, sql.NullString{})
	appendNumberedField(30, "Tgl Perceraian", session.TanggalPerceraian, sql.NullString{})
	appendNumberedField(31, "Cacat", session.CacatID, session.CacatNama)
	appendNumberedField(32, "Cara KB", session.CaraKbID, session.CaraKBNama)
	appendNumberedField(33, "Status Hamil", session.HamilID, session.HamilNama)
	appendNumberedField(34, "KTP Elektronik", session.KtpElID, session.KTPElNama)
	appendNumberedField(35, "Status Rekam", session.StatusRekamID, session.StatusRekamNama)
	appendNumberedField(36, "Alamat Sekarang", session.AlamatSekarang, sql.NullString{})
	appendNumberedField(37, "Status Dasar", session.StatusDasarID, session.StatusDasarNama)
	appendNumberedField(38, "Suku", session.SukuID, session.SukuNama)
	appendNumberedField(39, "Tag Card", session.TagCard, sql.NullString{})
	appendNumberedField(40, "Asuransi", session.IDAsuransiID, session.AsuransiNama)
	appendNumberedField(41, "No. Asuransi", session.NoAsuransi, sql.NullString{})

	return strings.TrimSpace(result.String()), nil
}

func FormatDate(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("02-01-2006")
}

// UpdateStepOnly updates the step of a session without modifying other fields.
func UpdateStepOnly(dbConn *sqlx.DB, jid string, step int) error {
	log.Printf("[DEBUG] Updating step to %d for jid: %s", step, jid)
	query := `
        UPDATE data_entry_sessions 
        SET current_step = $1,
            awaiting_answer = true,
            updated_at = NOW()
        WHERE jid = $2`
	_, err := dbConn.Exec(query, step, jid)
	return err
}

// Add these new functions
func SetEditField(dbConn *sqlx.DB, jid string, field string) error {
	_, err := dbConn.Exec(`UPDATE data_entry_sessions SET edit_field = $1 WHERE jid = $2`, field, jid)
	return err
}

func GetEditField(dbConn *sqlx.DB, jid string) (string, error) {
	var field string
	err := dbConn.Get(&field, `SELECT edit_field FROM data_entry_sessions WHERE jid = $1`, jid)
	return field, err
}

// Add new function to handle edit_field column
func EnsureEditFieldColumn(dbConn *sqlx.DB) error {
	// Check if column exists
	var exists bool
	err := dbConn.QueryRow(`
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_schema = 'public'
            AND table_name = 'data_entry_sessions' 
            AND column_name = 'edit_field'
        );
    `).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check edit_field column: %v", err)
	}

	// Add column if it doesn't exist
	if !exists {
		_, err = dbConn.Exec(`
            ALTER TABLE data_entry_sessions 
            ADD COLUMN edit_field text;
        `)
		if err != nil {
			return fmt.Errorf("failed to add edit_field column: %v", err)
		}
		log.Printf("[DEBUG] Added edit_field column to data_entry_sessions table")
	}

	return nil
}

// GetFieldValue gets a specific field value from data_entry_sessions
func GetFieldValue(dbConn *sqlx.DB, jid string, field string) (string, error) {
	var value string
	query := fmt.Sprintf("SELECT %s FROM data_entry_sessions WHERE jid = $1", field)
	err := dbConn.QueryRow(query, jid).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}
