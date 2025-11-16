package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"context"

	"github.com/jmoiron/sqlx"
)

type DataEntrySession struct {
	JID                  string         `db:"jid"`
	CurrentStep          int            `db:"current_step"`
	AwaitingAnswer       bool           `db:"awaiting_answer"`
	CurrentFlow        sql.NullString `db:"current_flow"`
	SheetRowNum          sql.NullInt64  `db:"sheet_row_num"`
	EditField            sql.NullString `db:"edit_field"`

	Dusun                sql.NullString `db:"dusun"`
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

	SuratValidNik      sql.NullString `db:"surat_valid_nik"`
	SuratFieldsPending sql.NullString `db:"surat_fields_pending"`
	SuratFieldNow      sql.NullString `db:"surat_field_now"`
	SuratDataMap       sql.NullString `db:"surat_data_map"`
	SuratTempAnswer    sql.NullString `db:"surat_temp_answer"`

	// Change lookup table name fields to use sql.NullString
	DusunNama            sql.NullString `db:"dusun_nama"`
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
	if err := EnsureColumn(dbConn, "data_entry_sessions", "edit_field", "TEXT"); err != nil { return nil, err }
	if err := EnsureColumn(dbConn, "data_entry_sessions", "sheet_row_num", "INTEGER"); err != nil { return nil, err }
	if err := EnsureColumn(dbConn, "data_entry_sessions", "current_flow", "TEXT"); err != nil { return nil, err }
	if err := EnsureSuratSessionColumns(dbConn); err != nil { return nil, err }

	var session DataEntrySession
	query := `SELECT * FROM data_entry_sessions WHERE jid = $1`
	err := dbConn.Get(&session, query, jid)
	
	if err == sql.ErrNoRows {
		log.Printf("[DEBUG] Creating new session for jid: %s", jid)
		newSession := DataEntrySession{
			JID:            jid,
			CurrentStep:    1,
			AwaitingAnswer: false,
			EditField:      sql.NullString{String: "", Valid: false},
		}
		_, err := dbConn.NamedExec(`
			INSERT INTO data_entry_sessions (jid, current_step, awaiting_answer, edit_field, current_flow) 
			VALUES (:jid, :current_step, :awaiting_answer, :edit_field, NULL)`, newSession)
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
	log.Printf("[DEBUG] Existing session found - Step: %d, Awaiting: %v, Flow: %s", session.CurrentStep, session.AwaitingAnswer, session.CurrentFlow.String)
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
						current_flow = 'FLOW_DATA_DIRI',

            dusun = NULL,
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
            nama_ayah = NULL,
            nama_ibu = NULL,
            status_dasar_id = NULL,
            suku_id = NULL,

            nik_ayah = NULL,
            nik_ibu = NULL,
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
            tag_card = NULL,
            id_asuransi_id = NULL,
            no_asuransi = NULL,

            edit_field = NULL,sheet_row_num = NULL,surat_valid_nik = NULL, surat_fields_pending = NULL, surat_field_now = NULL, surat_data_map = NULL,surat_temp_answer = NULL,
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

	session, err := GetFullSessionData(dbConn, jid)
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

	// Helper function untuk menentukan emoji berdasarkan field type
	getFieldEmoji := func(num int) string {
		emojiMap := map[int]string{
			1: "📍", 2: "🏘️", 3: "👤", 4: "🏠", 5: "🆔",
			6: "👨", 7: "🗺️", 8: "📅", 9: "⛪", 10: "🎓",
			11: "📚", 12: "💼", 13: "💍", 14: "👨‍👩‍👧‍👦", 15: "🌍",
			16: "👨‍🦱", 17: "👩", 18: "❤️", 19: "🎭", 20: "🆔",
			21: "🆔", 22: "🩸", 23: "📜", 24: "✈️", 25: "📅",
			26: "📋", 27: "📋", 28: "📅", 29: "📋", 30: "📅",
			31: "♿", 32: "🏥", 33: "🤰", 34: "🎫", 35: "📝",
			36: "🏠", 37: "🏷️", 38: "🏥", 39: "💳",
		}
		if emoji, ok := emojiMap[num]; ok {
			return emoji
		}
		return "•"
	}

	appendNumberedField := func(num int, label string, value interface{}, name sql.NullString) {
		emoji := getFieldEmoji(num)
		if str, ok := value.(sql.NullInt64); ok && str.Valid {
			if name.Valid && name.String != "" {
				fmt.Fprintf(&result, "%s *%d. %s:* %s\n", emoji, num, label, name.String)
			} else {
				fmt.Fprintf(&result, "%s *%d. %s:* %d\n", emoji, num, label, str.Int64)
			}
		} else if str, ok := value.(sql.NullString); ok && str.Valid {
			fmt.Fprintf(&result, "%s *%d. %s:* %s\n", emoji, num, label, str.String)
		} else if str, ok := value.(sql.NullTime); ok && str.Valid {
			fmt.Fprintf(&result, "%s *%d. %s:* %s\n", emoji, num, label, FormatDate(str))
		} else {
			fmt.Fprintf(&result, "%s *%d. %s:* -\n", emoji, num, label)
		}
	}

	// Always show all fields with their numbers and IDs
	appendNumberedField(1, "Dusun", session.Dusun, sql.NullString{})
	appendNumberedField(2, "RT", session.RT, sql.NullString{})
	appendNumberedField(3, "Nama", session.Nama, sql.NullString{})
	appendNumberedField(4, "No. KK", session.NoKK, sql.NullString{})
	appendNumberedField(5, "NIK", session.NIK, sql.NullString{})
	appendNumberedField(6, "Jenis Kelamin", session.SexID, session.SexNama)
	appendNumberedField(7, "Tempat Lahir", session.TempatLahir, sql.NullString{})
	appendNumberedField(8, "Tanggal Lahir", session.TanggalLahir, sql.NullString{})
	appendNumberedField(9, "Agama", session.AgamaID, session.AgamaNama)
	appendNumberedField(10, "Pendidikan KK", session.PendidikanKkID, session.PendidikanKKNama)
	appendNumberedField(11, "Pendidikan Sedang", session.PendidikanSedangID, session.PendidikanSedangNama)
	appendNumberedField(12, "Pekerjaan", session.PekerjaanID, session.PekerjaanNama)
	appendNumberedField(13, "Status Kawin", session.StatusKawinID, session.StatusKawinNama)
	appendNumberedField(14, "Level KK", session.KkLevelID, session.KKLevelNama)
	appendNumberedField(15, "Warganegara", session.WarganegaraID, session.WarganegaraNama)
	appendNumberedField(16, "Nama Ayah", session.NamaAyah, sql.NullString{})
	appendNumberedField(17, "Nama Ibu", session.NamaIbu, sql.NullString{})
	appendNumberedField(18, "Status Dasar", session.StatusDasarID, session.StatusDasarNama)
	appendNumberedField(19, "Suku", session.SukuID, session.SukuNama)
	appendNumberedField(20, "NIK Ayah", session.NikAyah, sql.NullString{})
	appendNumberedField(21, "NIK Ibu", session.NikIbu, sql.NullString{})
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
	appendNumberedField(37, "Tag Card", session.TagCard, sql.NullString{})
	appendNumberedField(38, "Asuransi", session.IDAsuransiID, session.AsuransiNama)
	appendNumberedField(39, "No. Asuransi", session.NoAsuransi, sql.NullString{})

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
	var field sql.NullString // Diubah ke NullString
	err := dbConn.Get(&field, `SELECT edit_field FROM data_entry_sessions WHERE jid = $1`, jid)
	return field.String, err // Kembalikan string (kosong jika NULL)
}

func EnsureSheetRowNumColumn(dbConn *sqlx.DB) error {
	var exists bool
	err := dbConn.QueryRow(`
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_schema = 'public'
            AND table_name = 'data_entry_sessions' 
            AND column_name = 'sheet_row_num'
        );
    `).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check sheet_row_num column: %v", err)
	}

	if !exists {
		_, err = dbConn.Exec(`
            ALTER TABLE data_entry_sessions 
            ADD COLUMN sheet_row_num INTEGER;
        `)
		if err != nil {
			return fmt.Errorf("failed to add sheet_row_num column: %v", err)
		}
		log.Printf("[DEBUG] Added sheet_row_num column to data_entry_sessions table")
	}
	return nil
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

func CheckNIKExists(dbConn *sqlx.DB, nik string, jid string) (bool, error) {
	var exists bool
	// Query ini mencari NIK di semua sesi, KECUALI sesi milik user saat ini
	query := "SELECT EXISTS(SELECT 1 FROM data_entry_sessions WHERE nik = $1 AND jid != $2)"
	err := dbConn.Get(&exists, query, nik, jid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func EnsureSuratSessionColumns(dbConn *sqlx.DB) error {
	columns := []string{
		"surat_valid_nik",
		"surat_fields_pending",
		"surat_field_now",
		"surat_data_map",
		"surat_temp_answer",
	}
	
	for _, col := range columns {
		var exists bool
		query := fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_schema = 'public'
				AND table_name = 'data_entry_sessions' 
				AND column_name = '%s'
			);
		`, col)
		err := dbConn.QueryRow(query).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check %s column: %v", col, err)
		}

		if !exists {
			log.Printf("[DEBUG] Column '%s' not found, adding it...", col)
			_, err = dbConn.Exec(fmt.Sprintf(`
				ALTER TABLE data_entry_sessions 
				ADD COLUMN %s TEXT;
			`, col))
			if err != nil {
				return fmt.Errorf("failed to add %s column: %v", col, err)
			}
			log.Printf("[DEBUG] Added %s column to data_entry_sessions table", col)
		}
	}
	return nil
}

func UpdateSessionField(dbConn *sqlx.DB, jid string, field string, value interface{}) error {
	query := fmt.Sprintf("UPDATE data_entry_sessions SET %s = $1, updated_at = NOW() WHERE jid = $2", field)
	_, err := dbConn.Exec(query, value, jid)
	if err != nil {
		log.Printf("[ERROR] Gagal UpdateSessionField (%s): %v", field, err)
	}
	return err
}

func GetSessionField(dbConn *sqlx.DB, jid string, field string) (sql.NullString, error) {
	var result sql.NullString
	query := fmt.Sprintf("SELECT %s FROM data_entry_sessions WHERE jid = $1", field)
	err := dbConn.Get(&result, query, jid)
	return result, err
}

func UpdateSessionFlow(dbConn *sqlx.DB, jid string, flow string, step int) error {
	log.Printf("[DEBUG] Updating flow to %s, step %d for jid: %s", flow, step, jid)
	query := `
        UPDATE data_entry_sessions 
        SET current_flow = $1,
            current_step = $2,
            awaiting_answer = true,
            updated_at = NOW()
        WHERE jid = $3`
	_, err := dbConn.Exec(query, flow, step, jid)
	return err
}

func EnsureFlowColumn(dbConn *sqlx.DB) error {
	return EnsureColumn(dbConn, "data_entry_sessions", "current_flow", "TEXT")
}

// (Helper EnsureColumn dari schema.go dipindah ke sini)
func EnsureColumn(dbConn *sqlx.DB, tableName, columnName, columnType string) error {
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

	if !exists {
		log.Printf("[DEBUG] Column '%s' not found in table '%s', adding it...", columnName, tableName)
		_, err = dbConn.Exec(fmt.Sprintf(`
			ALTER TABLE %s 
			ADD COLUMN %s %s;
		`, tableName, columnName, columnType))
		if err != nil {
			return fmt.Errorf("failed to add %s column: %v", columnName, err)
		}
		log.Printf("[DEBUG] Added %s column to %s table", columnName, tableName)
	}
	return nil
}

func GetSessionByJID(db *sqlx.DB, jid string) (*DataEntrySession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := &DataEntrySession{}
	query := `SELECT * FROM data_entry_sessions WHERE jid = $1`
	err := db.GetContext(ctx, session, query, jid)
	if err != nil {
		return nil, err
	}
	return session, nil
}
