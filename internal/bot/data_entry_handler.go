package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

var steps = map[int]struct {
	Question string
	Field    string
	IsInt    bool
	IsDate   bool
	Options  map[int]string
}{
	1:  {Question: "Masukkan Alamat:", Field: "alamat"},
	2:  {Question: "Masukkan Dusun:", Field: "dusun"},
	3:  {Question: "Masukkan RW:", Field: "rw"},
	4:  {Question: "Masukkan RT:", Field: "rt"},
	5:  {Question: "Masukkan Nama:", Field: "nama"},
	6:  {Question: "Masukkan No. KK:", Field: "no_kk"},
	7:  {Question: "Masukkan NIK:", Field: "nik"},
	8:  {Question: "Pilih Jenis Kelamin:", Field: "sex_id", IsInt: true, Options: map[int]string{1: "LAKI-LAKI", 2: "PEREMPUAN"}},
	9:  {Question: "Masukkan Tempat Lahir:", Field: "tempat_lahir"},
	10: {Question: "Masukkan Tanggal Lahir (YYYY-MM-DD):", Field: "tanggal_lahir", IsDate: true},
	11: {Question: "Masukkan ID Agama (agama_id):", Field: "agama_id", IsInt: true},
	12: {Question: "Masukkan ID Pendidikan KK (pendidikan_kk_id):", Field: "pendidikan_kk_id", IsInt: true},
	13: {Question: "Masukkan ID Pendidikan Sedang (pendidikan_sedang_id):", Field: "pendidikan_sedang_id", IsInt: true},
	14: {Question: "Masukkan ID Pekerjaan (pekerjaan_id):", Field: "pekerjaan_id", IsInt: true},
	15: {Question: "Masukkan ID Status Kawin (status_kawin_id):", Field: "status_kawin_id", IsInt: true},
	16: {Question: "Masukkan ID Level KK (kk_level_id):", Field: "kk_level_id", IsInt: true},
	17: {Question: "Masukkan ID Warganegara (warganegara_id):", Field: "warganegara_id", IsInt: true},
	18: {Question: "Masukkan NIK Ayah:", Field: "nik_ayah"},
	19: {Question: "Masukkan Nama Ayah:", Field: "nama_ayah"},
	20: {Question: "Masukkan NIK Ibu:", Field: "nik_ibu"},
	21: {Question: "Masukkan Nama Ibu:", Field: "nama_ibu"},
	22: {Question: "Masukkan ID Golongan Darah (golongan_darah_id):", Field: "golongan_darah_id", IsInt: true},
	23: {Question: "Masukkan No. Akta Lahir:", Field: "akta_lahir"},
	24: {Question: "Masukkan No. Dokumen Paspor:", Field: "dokumen_passport"},
	25: {Question: "Masukkan Tanggal Akhir Paspor (YYYY-MM-DD):", Field: "tanggal_akhir_passport", IsDate: true},
	26: {Question: "Masukkan No. Dokumen KITAS:", Field: "dokumen_kitas"},
	27: {Question: "Masukkan No. Akta Perkawinan:", Field: "akta_perkawinan"},
	28: {Question: "Masukkan Tanggal Perkawinan (YYYY-MM-DD):", Field: "tanggal_perkawinan", IsDate: true},
	29: {Question: "Masukkan No. Akta Perceraian:", Field: "akta_perceraian"},
	30: {Question: "Masukkan Tanggal Perceraian (YYYY-MM-DD):", Field: "tanggal_perceraian", IsDate: true},
	31: {Question: "Masukkan ID Cacat (cacat_id):", Field: "cacat_id", IsInt: true},
	32: {Question: "Masukkan ID Cara KB (cara_kb_id):", Field: "cara_kb_id", IsInt: true},
	33: {Question: "Masukkan ID Status Hamil (hamil_id):", Field: "hamil_id", IsInt: true},
	34: {Question: "Masukkan ID KTP Elektronik (ktp_el_id):", Field: "ktp_el_id", IsInt: true},
	35: {Question: "Masukkan ID Status Rekam (status_rekam_id):", Field: "status_rekam_id", IsInt: true},
	36: {Question: "Masukkan Alamat Sekarang:", Field: "alamat_sekarang"},
	37: {Question: "Masukkan ID Status Dasar (status_dasar_id):", Field: "status_dasar_id", IsInt: true},
	38: {Question: "Masukkan ID Suku (suku_id):", Field: "suku_id", IsInt: true},
	39: {Question: "Masukkan Tag Card:", Field: "tag_card"},
	40: {Question: "Masukkan ID Asuransi (id_asuransi_id):", Field: "id_asuransi_id", IsInt: true},
	41: {Question: "Masukkan No. Asuransi:", Field: "no_asuransi"},
}

func HandleDataEntry(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession) string {
	log.Printf("[DEBUG] Handling data entry for step %d", session.CurrentStep)

	stepInfo, ok := steps[session.CurrentStep]
	if !ok {
		log.Printf("[DEBUG] No more steps available, completing session")
		err := db.DeleteDataEntrySession(dbConn, jid)
		if err != nil {
			log.Printf("[ERROR] Failed to delete completed session: %v", err)
		}
		return "Terima kasih! Semua data telah berhasil dimasukkan."
	}

	// Validate input
	value, err := validateInput(text, stepInfo)
	if err != nil {
		log.Printf("[DEBUG] Input validation failed: %v", err)
		if stepInfo.Options != nil {
			return formatQuestionWithOptions(stepInfo.Question, stepInfo.Options)
		}
		return stepInfo.Question
	}

	log.Printf("[DEBUG] Updating session with valid input for step %d", session.CurrentStep)
	// Update session with valid input
	if err := db.UpdateDataEntrySession(dbConn, jid, stepInfo.Field, value); err != nil {
		log.Printf("Error updating session: %v", err)
		return "Maaf, terjadi kesalahan sistem."
	}

	// Get next question
	nextStep := session.CurrentStep + 1
	if nextStepInfo, ok := steps[nextStep]; ok {
		return formatQuestion(nextStepInfo)
	}

	// Finish session
	if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
		log.Printf("Error deleting session: %v", err)
	}
	return "Terima kasih! Semua data telah berhasil dimasukkan."
}

func validateInput(text string, step struct {
	Question string
	Field    string
	IsInt    bool
	IsDate   bool
	Options  map[int]string
}) (interface{}, error) {
	if step.Options != nil {
		choice, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("nomor pilihan tidak valid")
		}
		if _, valid := step.Options[choice]; !valid {
			return nil, fmt.Errorf("pilihan tidak valid")
		}
		return choice, nil
	} else if step.IsInt {
		value, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("input tidak valid, harap masukkan angka")
		}
		return value, nil
	} else if step.IsDate {
		value, err := time.Parse("2006-01-02", text)
		if err != nil {
			return nil, fmt.Errorf("format tanggal tidak valid, harap gunakan format YYYY-MM-DD")
		}
		return value, nil
	}
	return text, nil
}

func formatQuestion(step struct {
	Question string
	Field    string
	IsInt    bool
	IsDate   bool
	Options  map[int]string
}) string {
	if step.Options != nil {
		return formatQuestionWithOptions(step.Question, step.Options)
	}
	return step.Question
}

func formatQuestionWithOptions(question string, options map[int]string) string {
	var builder strings.Builder
	builder.WriteString(question)
	for id, name := range options {
		builder.WriteString(fmt.Sprintf("\n%d. %s", id, name))
	}
	return builder.String()
}
