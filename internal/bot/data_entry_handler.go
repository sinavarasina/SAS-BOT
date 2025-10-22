package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

// Type definitions
type SexOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type AgamaOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type SexData struct {
	Sex []SexOption `json:"sex"`
}

type AgamaData struct {
	Agama []AgamaOption `json:"agama"`
}

type PendidikanKKOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type PendidikanSedangOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type PekerjaanOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type StatusKawinOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type KKLevelOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type WarganegaraOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type GolonganDarahOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type CacatOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type CaraKBOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type HamilOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type KTPElektronikOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type StatusRekamOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type StatusDasarOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type SukuOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type AsuransiOption struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

// Add new data structures
type PendidikanKKData struct {
	PendidikanKK []PendidikanKKOption `json:"pendidikan_kk"`
}

type PendidikanSedangData struct {
	PendidikanSedang []PendidikanSedangOption `json:"pendidikan_sedang"`
}

type PekerjaanData struct {
	Pekerjaan []PekerjaanOption `json:"pekerjaan"`
}

type StatusKawinData struct {
	StatusKawin []StatusKawinOption `json:"status_kawin"`
}

type KKLevelData struct {
	KKLevel []KKLevelOption `json:"kk_level"`
}

type WarganegaraData struct {
	Warganegara []WarganegaraOption `json:"warganegara"`
}

type GolonganDarahData struct {
	GolonganDarah []GolonganDarahOption `json:"golongan_darah"`
}

type CacatData struct {
	Cacat []CacatOption `json:"cacat"`
}

type CaraKBData struct {
	CaraKB []CaraKBOption `json:"cara_kb"`
}

type HamilData struct {
	Hamil []HamilOption `json:"hamil"`
}

type KTPElektronikData struct {
	KTPElektronik []KTPElektronikOption `json:"ktp_el"`
}

type StatusRekamData struct {
	StatusRekam []StatusRekamOption `json:"status_rekam"`
}

type StatusDasarData struct {
	StatusDasar []StatusDasarOption `json:"status_dasar"`
}

type SukuData struct {
	Suku []SukuOption `json:"suku"`
}

type AsuransiData struct {
	Asuransi []AsuransiOption `json:"asuransi"`
}

// Define step structure type
type Step struct {
	Question string
	Field    string
	IsInt    bool
	IsDate   bool
	Options  map[int]string
}

// All package-level variables need to be declared before functions
var (
	sexOptions, agamaOptions, pendidikanKKOptions, pendidikanSedangOptions, pekerjaanOptions, statusKawinOptions, kkLevelOptions, warganegaraOptions, golonganDarahOptions, cacatOptions, caraKBOptions, hamilOptions, ktpElOptions, statusRekamOptions, statusDasarOptions, sukuOptions, asuransiOptions = loadJSONOptions()
	steps                                                                                                                                                                                                                                                                                                 = map[int]Step{
		1:  {"Masukkan Alamat:", "alamat", false, false, nil},
		2:  {"Masukkan Dusun:", "dusun", false, false, nil},
		3:  {"Masukkan RW:", "rw", false, false, nil},
		4:  {"Masukkan RT:", "rt", false, false, nil},
		5:  {"Masukkan Nama:", "nama", false, false, nil},
		6:  {"Masukkan No. KK:", "no_kk", false, false, nil},
		7:  {"Masukkan NIK:", "nik", false, false, nil},
		8:  {"Pilih Jenis Kelamin:", "sex_id", true, false, sexOptions},
		9:  {"Masukkan Tempat Lahir:", "tempat_lahir", false, false, nil},
		10: {"Masukkan Tanggal Lahir (DD-MM-YYYY):", "tanggal_lahir", false, true, nil},
		11: {"Pilih Agama:", "agama_id", true, false, agamaOptions},
		12: {"Pilih Pendidikan KK:", "pendidikan_kk_id", true, false, pendidikanKKOptions},
		13: {"Pilih Pendidikan Sedang:", "pendidikan_sedang_id", true, false, pendidikanSedangOptions},
		14: {"Pilih Pekerjaan:", "pekerjaan_id", true, false, pekerjaanOptions},
		15: {"Pilih Status Kawin:", "status_kawin_id", true, false, statusKawinOptions},
		16: {"Pilih Level KK:", "kk_level_id", true, false, kkLevelOptions},
		17: {"Pilih Warganegara:", "warganegara_id", true, false, warganegaraOptions},
		18: {"Masukkan NIK Ayah:", "nik_ayah", false, false, nil},
		19: {"Masukkan Nama Ayah:", "nama_ayah", false, false, nil},
		20: {"Masukkan NIK Ibu:", "nik_ibu", false, false, nil},
		21: {"Masukkan Nama Ibu:", "nama_ibu", false, false, nil},
		22: {"Pilih Golongan Darah:", "golongan_darah_id", true, false, golonganDarahOptions},
		23: {"Masukkan No. Akta Lahir:", "akta_lahir", false, false, nil},
		24: {"Masukkan No. Dokumen Paspor:", "dokumen_passport", false, false, nil},
		25: {"Masukkan Tanggal Akhir Paspor (DD-MM-YYYY):", "tanggal_akhir_passport", false, true, nil},
		26: {"Masukkan No. Dokumen KITAS:", "dokumen_kitas", false, false, nil},
		27: {"Masukkan No. Akta Perkawinan:", "akta_perkawinan", false, false, nil},
		28: {"Masukkan Tanggal Perkawinan (DD-MM-YYYY):", "tanggal_perkawinan", false, true, nil},
		29: {"Masukkan No. Akta Perceraian:", "akta_perceraian", false, false, nil},
		30: {"Masukkan Tanggal Perceraian (DD-MM-YYYY):", "tanggal_perceraian", false, true, nil},
		31: {"Pilih Cacat:", "cacat_id", true, false, cacatOptions},
		32: {"Pilih Cara KB:", "cara_kb_id", true, false, caraKBOptions},
		33: {"Pilih Status Hamil:", "hamil_id", true, false, hamilOptions},
		34: {"Pilih KTP Elektronik:", "ktp_el_id", true, false, ktpElOptions},
		35: {"Pilih Status Rekam:", "status_rekam_id", true, false, statusRekamOptions},
		36: {"Masukkan Alamat Sekarang:", "alamat_sekarang", false, false, nil},
		37: {"Pilih Status Dasar:", "status_dasar_id", true, false, statusDasarOptions},
		38: {"Pilih Suku:", "suku_id", true, false, sukuOptions},
		39: {"Masukkan Tag Card:", "tag_card", false, false, nil},
		40: {"Pilih Asuransi:", "id_asuransi_id", true, false, asuransiOptions},
		41: {"Masukkan No. Asuransi:", "no_asuransi", false, false, nil},
	}
)

func loadJSONOptions() (map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string, map[int]string) {
	sexOptions := make(map[int]string)
	agamaOptions := make(map[int]string)
	pendidikanKKOptions := make(map[int]string)
	pendidikanSedangOptions := make(map[int]string)
	pekerjaanOptions := make(map[int]string)
	statusKawinOptions := make(map[int]string)
	kkLevelOptions := make(map[int]string)
	warganegaraOptions := make(map[int]string)
	golonganDarahOptions := make(map[int]string)
	cacatOptions := make(map[int]string)
	caraKBOptions := make(map[int]string)
	hamilOptions := make(map[int]string)
	ktpElOptions := make(map[int]string)
	statusRekamOptions := make(map[int]string)
	statusDasarOptions := make(map[int]string)
	sukuOptions := make(map[int]string)
	asuransiOptions := make(map[int]string)

	// Load sex options
	sexFile, err := os.ReadFile(filepath.Join("json", "8_sex.json"))
	if err == nil {
		var sexData SexData
		if err := json.Unmarshal(sexFile, &sexData); err == nil {
			for _, opt := range sexData.Sex {
				sexOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load agama options
	agamaFile, err := os.ReadFile(filepath.Join("json", "11_agama.json"))
	if err == nil {
		var agamaData AgamaData
		if err := json.Unmarshal(agamaFile, &agamaData); err == nil {
			for _, opt := range agamaData.Agama {
				agamaOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pendidikan_kk options
	if data, err := os.ReadFile(filepath.Join("json", "12_pendidikan_kk.json")); err == nil {
		var pendidikanKKData PendidikanKKData
		if err := json.Unmarshal(data, &pendidikanKKData); err == nil {
			for _, opt := range pendidikanKKData.PendidikanKK {
				pendidikanKKOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pendidikan_sedang options
	if data, err := os.ReadFile(filepath.Join("json", "13_pendidikan_sedang.json")); err == nil {
		var pendidikanSedangData PendidikanSedangData
		if err := json.Unmarshal(data, &pendidikanSedangData); err == nil {
			for _, opt := range pendidikanSedangData.PendidikanSedang {
				pendidikanSedangOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pekerjaan options
	if data, err := os.ReadFile(filepath.Join("json", "14_pekerjaan.json")); err == nil {
		var pekerjaanData PekerjaanData
		if err := json.Unmarshal(data, &pekerjaanData); err == nil {
			for _, opt := range pekerjaanData.Pekerjaan {
				pekerjaanOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_kawin options
	if data, err := os.ReadFile(filepath.Join("json", "15_status_kawin.json")); err == nil {
		var statusKawinData StatusKawinData
		if err := json.Unmarshal(data, &statusKawinData); err == nil {
			for _, opt := range statusKawinData.StatusKawin {
				statusKawinOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load kk_level options
	if data, err := os.ReadFile(filepath.Join("json", "16_kk_level.json")); err == nil {
		var kkLevelData KKLevelData
		if err := json.Unmarshal(data, &kkLevelData); err == nil {
			for _, opt := range kkLevelData.KKLevel {
				kkLevelOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load warganegara options
	if data, err := os.ReadFile(filepath.Join("json", "17_warganegara.json")); err == nil {
		var warganegaraData WarganegaraData
		if err := json.Unmarshal(data, &warganegaraData); err == nil {
			for _, opt := range warganegaraData.Warganegara {
				warganegaraOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load golongan_darah options
	if data, err := os.ReadFile(filepath.Join("json", "22_golongan_darah.json")); err == nil {
		var golonganDarahData GolonganDarahData
		if err := json.Unmarshal(data, &golonganDarahData); err == nil {
			for _, opt := range golonganDarahData.GolonganDarah {
				golonganDarahOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load cacat options
	if data, err := os.ReadFile(filepath.Join("json", "31_cacat.json")); err == nil {
		var cacatData CacatData
		if err := json.Unmarshal(data, &cacatData); err == nil {
			for _, opt := range cacatData.Cacat {
				cacatOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load cara_kb options
	if data, err := os.ReadFile(filepath.Join("json", "32_cara_kb.json")); err == nil {
		var caraKBData CaraKBData
		if err := json.Unmarshal(data, &caraKBData); err == nil {
			for _, opt := range caraKBData.CaraKB {
				caraKBOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load hamil options
	if data, err := os.ReadFile(filepath.Join("json", "33_hamil.json")); err == nil {
		var hamilData HamilData
		if err := json.Unmarshal(data, &hamilData); err == nil {
			for _, opt := range hamilData.Hamil {
				hamilOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load ktp_el options
	if data, err := os.ReadFile(filepath.Join("json", "34_ktp_el.json")); err == nil {
		var ktpElData KTPElektronikData
		if err := json.Unmarshal(data, &ktpElData); err == nil {
			for _, opt := range ktpElData.KTPElektronik {
				ktpElOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_rekam options
	if data, err := os.ReadFile(filepath.Join("json", "35_status_rekam.json")); err == nil {
		var statusRekamData StatusRekamData
		if err := json.Unmarshal(data, &statusRekamData); err == nil {
			for _, opt := range statusRekamData.StatusRekam {
				statusRekamOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_dasar options
	if data, err := os.ReadFile(filepath.Join("json", "37_status_dasar.json")); err == nil {
		var statusDasarData StatusDasarData
		if err := json.Unmarshal(data, &statusDasarData); err == nil {
			for _, opt := range statusDasarData.StatusDasar {
				statusDasarOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load suku options
	if data, err := os.ReadFile(filepath.Join("json", "38_suku.json")); err == nil {
		var sukuData SukuData
		if err := json.Unmarshal(data, &sukuData); err == nil {
			for _, opt := range sukuData.Suku {
				sukuOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load asuransi options
	if data, err := os.ReadFile(filepath.Join("json", "40_asuransi.json")); err == nil {
		var asuransiData AsuransiData
		if err := json.Unmarshal(data, &asuransiData); err == nil {
			for _, opt := range asuransiData.Asuransi {
				asuransiOptions[opt.ID] = opt.Nama
			}
		}
	}

	return sexOptions, agamaOptions, pendidikanKKOptions, pendidikanSedangOptions, pekerjaanOptions, statusKawinOptions, kkLevelOptions, warganegaraOptions, golonganDarahOptions, cacatOptions, caraKBOptions, hamilOptions, ktpElOptions, statusRekamOptions, statusDasarOptions, sukuOptions, asuransiOptions
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

func validateInput(text string, step Step) (interface{}, error) {
	if step.Options != nil {
		choice, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("nomor pilihan tidak valid\n\nPilihan yang tersedia:")
		}
		if _, valid := step.Options[choice]; !valid {
			optionsStr := "\n\nPilihan yang tersedia:"
			ids := make([]int, 0, len(step.Options))
			for id := range step.Options {
				ids = append(ids, id)
			}
			sort.Ints(ids)
			for _, id := range ids {
				optionsStr += fmt.Sprintf("\n%d. %s", id, step.Options[id])
			}
			return nil, fmt.Errorf("pilihan tidak valid%s", optionsStr)
		}
		return choice, nil
	} else if step.IsInt {
		value, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("input tidak valid, harap masukkan angka")
		}
		return value, nil
	} else if step.IsDate {
		value, err := time.Parse("02-01-2006", text)
		if err != nil {
			return nil, fmt.Errorf("format tanggal tidak valid, harap gunakan format DD-MM-YYYY")
		}
		return value, nil
	}
	return text, nil
}

func formatQuestion(step Step) string {
	if step.Options != nil {
		return formatQuestionWithOptions(step.Question, step.Options)
	}
	return step.Question
}

func formatQuestionWithOptions(question string, options map[int]string) string {
	var builder strings.Builder
	builder.WriteString(question)

	// Create sorted slice of IDs
	ids := make([]int, 0, len(options))
	for id := range options {
		ids = append(ids, id)
	}
	// Sort IDs in ascending order
	sort.Ints(ids)

	// Build output using sorted IDs
	for _, id := range ids {
		builder.WriteString(fmt.Sprintf("\n%d. %s", id, options[id]))
	}
	return builder.String()
}
