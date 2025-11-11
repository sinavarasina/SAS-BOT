// Gw pengen ini ntar di pecah jadi module baru di dalam folder bot/

package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
	"github.com/sinavarasina/SAS-BOT/internal/surat"
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
	steps = map[int]Step{
		1:  {"Masukkan Dusun:", "dusun", false, false, nil},
		2:  {"Masukkan RT:", "rt", false, false, nil},
		3:  {"Masukkan Nama:", "nama", false, false, nil},
		4:  {"Masukkan No. KK:", "no_kk", false, false, nil},
		5:  {"Masukkan NIK:", "nik", false, false, nil},
		6:  {"Pilih Jenis Kelamin:", "sex_id", true, false, sexOptions},
		7:  {"Masukkan Tempat Lahir:", "tempat_lahir", false, false, nil},
		8:  {"Masukkan Tanggal Lahir (DD-MM-YYYY):", "tanggal_lahir", false, true, nil},
		9:  {"Pilih Agama:", "agama_id", true, false, agamaOptions},
		10: {"Pilih Pendidikan Dalam KK:", "pendidikan_kk_id", true, false, pendidikanKKOptions},
		11: {"Pilih Pendidikan Sedang Ditempuh:", "pendidikan_sedang_id", true, false, pendidikanSedangOptions},
		12: {"Pilih Pekerjaan:", "pekerjaan_id", true, false, pekerjaanOptions},
		13: {"Pilih Status Kawin:", "status_kawin_id", true, false, statusKawinOptions},
		14: {"Pilih Status Hubungan Dalam KK:", "kk_level_id", true, false, kkLevelOptions},
		15: {"Pilih Warganegara:", "warganegara_id", true, false, warganegaraOptions},
		16: {"Masukkan Nama Ayah:", "nama_ayah", false, false, nil},
		17: {"Masukkan Nama Ibu:", "nama_ibu", false, false, nil},
		18: {"Pilih Status Dasar:", "status_dasar_id", true, false, statusDasarOptions},
		19: {"Pilih Suku:", "suku_id", true, false, sukuOptions}, // Checkpoint baru
	}
)

const (
	STEP_START                  = 1
	STEP_NIK                    = 5
	STEP_DATA_INTI              = 19  // Tetap 19 (Suku), bukan 17
	STEP_GOLONGAN_DARAH         = 22
	STEP_CONFIRMATION           = 40
	STEP_EDIT                   = 41
	STEP_CHECKPOINT_DATA_INTI   = 100
	STEP_NIK_DUPLICATE          = 101
	STEP_MENU_DATA_DIRI         = 200 // Sesi menunggu Pilihan (Input, Edit, Hapus)
	STEP_EDIT_CARI_NIK          = 201 // Sesi menunggu NIK untuk di-edit
	STEP_PENGADUAN_WAITING      = 301
	STEP_PENGADUAN_MENU         = 300
	STEP_PENGADUAN_CARI_ID      = 302
	STEP_PENGADUAN_VALIDASI_NIK = 303
	STEP_SURAT_MENU_UTAMA       = 500 // Menunggu (1. Ajukan / 2. Cek)
	STEP_SURAT_VALIDASI_NIK     = 501 // Menunggu NIK untuk Ajuan
	STEP_SURAT_PILIH_JENIS      = 502 // Menunggu (1-5)
	STEP_SURAT_INPUT_DATA       = 503 // Menunggu jawaban field tambahan
	STEP_SURAT_KONFIRMASI_FIELD = 504 // Menunggu (ya/edit)
	STEP_SURAT_CEK_PROGRES      = 505 // Menunggu ID Unik untuk Cek
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

	// Load sex options (6_sex.json)
	if data, err := os.ReadFile(filepath.Join("json", "6_sex.json")); err == nil {
		var sexData SexData
		if err := json.Unmarshal(data, &sexData); err == nil {
			for _, opt := range sexData.Sex {
				sexOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load agama options (9_agama.json)
	if data, err := os.ReadFile(filepath.Join("json", "9_agama.json")); err == nil {
		var agamaData AgamaData
		if err := json.Unmarshal(data, &agamaData); err == nil {
			for _, opt := range agamaData.Agama {
				agamaOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pendidikan_kk options (10_pendidikan_kk.json)
	if data, err := os.ReadFile(filepath.Join("json", "10_pendidikan_kk.json")); err == nil {
		var pendidikanKKData PendidikanKKData
		if err := json.Unmarshal(data, &pendidikanKKData); err == nil {
			for _, opt := range pendidikanKKData.PendidikanKK {
				pendidikanKKOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pendidikan_sedang options (11_pendidikan_sedang.json)
	if data, err := os.ReadFile(filepath.Join("json", "11_pendidikan_sedang.json")); err == nil {
		var pendidikanSedangData PendidikanSedangData
		if err := json.Unmarshal(data, &pendidikanSedangData); err == nil {
			for _, opt := range pendidikanSedangData.PendidikanSedang {
				pendidikanSedangOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load pekerjaan options (12_pekerjaan.json)
	if data, err := os.ReadFile(filepath.Join("json", "12_pekerjaan.json")); err == nil {
		var pekerjaanData PekerjaanData
		if err := json.Unmarshal(data, &pekerjaanData); err == nil {
			for _, opt := range pekerjaanData.Pekerjaan {
				pekerjaanOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_kawin options (13_status_kawin.json)
	if data, err := os.ReadFile(filepath.Join("json", "13_status_kawin.json")); err == nil {
		var statusKawinData StatusKawinData
		if err := json.Unmarshal(data, &statusKawinData); err == nil {
			for _, opt := range statusKawinData.StatusKawin {
				statusKawinOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load kk_level options (14_kk_level.json)
	if data, err := os.ReadFile(filepath.Join("json", "14_kk_level.json")); err == nil {
		var kkLevelData KKLevelData
		if err := json.Unmarshal(data, &kkLevelData); err == nil {
			for _, opt := range kkLevelData.KKLevel {
				kkLevelOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load warganegara options (15_warganegara.json)
	if data, err := os.ReadFile(filepath.Join("json", "15_warganegara.json")); err == nil {
		var warganegaraData WarganegaraData
		if err := json.Unmarshal(data, &warganegaraData); err == nil {
			for _, opt := range warganegaraData.Warganegara {
				warganegaraOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load golongan_darah options (22_golongan_darah.json)
	if data, err := os.ReadFile(filepath.Join("json", "22_golongan_darah.json")); err == nil {
		var golonganDarahData GolonganDarahData
		if err := json.Unmarshal(data, &golonganDarahData); err == nil {
			for _, opt := range golonganDarahData.GolonganDarah {
				golonganDarahOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load cacat options (31_cacat.json)
	if data, err := os.ReadFile(filepath.Join("json", "31_cacat.json")); err == nil {
		var cacatData CacatData
		if err := json.Unmarshal(data, &cacatData); err == nil {
			for _, opt := range cacatData.Cacat {
				cacatOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load cara_kb options (32_cara_kb.json)
	if data, err := os.ReadFile(filepath.Join("json", "32_cara_kb.json")); err == nil {
		var caraKBData CaraKBData
		if err := json.Unmarshal(data, &caraKBData); err == nil {
			for _, opt := range caraKBData.CaraKB {
				caraKBOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load hamil options (33_hamil.json)
	if data, err := os.ReadFile(filepath.Join("json", "33_hamil.json")); err == nil {
		var hamilData HamilData
		if err := json.Unmarshal(data, &hamilData); err == nil {
			for _, opt := range hamilData.Hamil {
				hamilOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load ktp_el options (34_ktp_el.json)
	if data, err := os.ReadFile(filepath.Join("json", "34_ktp_el.json")); err == nil {
		var ktpElData KTPElektronikData
		if err := json.Unmarshal(data, &ktpElData); err == nil {
			for _, opt := range ktpElData.KTPElektronik {
				ktpElOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_rekam options (35_status_rekam.json)
	if data, err := os.ReadFile(filepath.Join("json", "35_status_rekam.json")); err == nil {
		var statusRekamData StatusRekamData
		if err := json.Unmarshal(data, &statusRekamData); err == nil {
			for _, opt := range statusRekamData.StatusRekam {
				statusRekamOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load status_dasar options (18_status_dasar.json)
	if data, err := os.ReadFile(filepath.Join("json", "18_status_dasar.json")); err == nil {
		var statusDasarData StatusDasarData
		if err := json.Unmarshal(data, &statusDasarData); err == nil {
			for _, opt := range statusDasarData.StatusDasar {
				statusDasarOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load suku options (19_suku.json)
	if data, err := os.ReadFile(filepath.Join("json", "19_suku.json")); err == nil {
		var sukuData SukuData
		if err := json.Unmarshal(data, &sukuData); err == nil {
			for _, opt := range sukuData.Suku {
				sukuOptions[opt.ID] = opt.Nama
			}
		}
	}

	// Load asuransi options (38_asuransi.json)
	if data, err := os.ReadFile(filepath.Join("json", "38_asuransi.json")); err == nil {
		var asuransiData AsuransiData
		if err := json.Unmarshal(data, &asuransiData); err == nil {
			for _, opt := range asuransiData.Asuransi {
				asuransiOptions[opt.ID] = opt.Nama
			}
		}
	}

	return sexOptions, agamaOptions, pendidikanKKOptions, pendidikanSedangOptions, pekerjaanOptions, statusKawinOptions, kkLevelOptions, warganegaraOptions, golonganDarahOptions, cacatOptions, caraKBOptions, hamilOptions, ktpElOptions, statusRekamOptions, statusDasarOptions, sukuOptions, asuransiOptions
}

func HandleDataEntry(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client) []string {
	log.Printf("[DEBUG] Handling data entry for step %d with input: '%s'", session.CurrentStep, text)

	switch session.CurrentStep {

	// --- 1. ALUR MENU UTAMA & SUB-MENU ---
	case STEP_MENU_DATA_DIRI: // User ada di sub-menu (200)
		switch text {
		case "1": // 1. Input Data Diri
			if err := db.StartNewSession(dbConn, jid); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{formatQuestion(steps[STEP_START])}
		case "2": // 2. Edit Data Diri
			if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT_CARI_NIK); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"🔍 Silakan masukkan *NIK 16 digit* yang datanya ingin Anda edit:"}
		default:
			return []string{"❌ Pilihan tidak valid. Silakan pilih 1 atau 2, atau ketik 'reset'."}
		}

	// --- 2. ALUR FITUR EDIT ---
	case STEP_EDIT_CARI_NIK: // Menunggu NIK untuk di-edit (201)
		data, err := db.GetDataPendudukByNIK(dbConn, text)
		if err != nil {
			log.Printf("[DEBUG] NIK %s tidak ditemukan di DB: %v", text, err)
			return []string{"❌ NIK tidak ditemukan di database. Silakan coba lagi atau ketik 'reset'."}
		}
		// Temukan! Salin data permanen ke sesi sementara
		if err := db.LoadSessionFromPenduduk(dbConn, jid, *data); err != nil {
			log.Printf("[ERROR] Gagal LoadSessionFromPenduduk: %v", err)
			return []string{"❌ NIK ditemukan, tapi gagal memuat data ke sesi. Hubungi admin."}
		}
		// Langsung lompat ke langkah konfirmasi (42)
		dataStr, _ := db.GetFormattedSessionData(dbConn, jid)
		return []string{"✅ Data ditemukan. Silakan periksa:\n\n📝 Ketik 'valid' untuk menyimpan atau 'edit' untuk mengubah data\n\n" + dataStr}

	// Menu 2
	case STEP_SURAT_MENU_UTAMA: // (Langkah 500)
		switch text {
		case "1": // 1. Ajukan Surat
			if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_VALIDASI_NIK); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Baik, silakan masukkan **NIK 16 digit** Anda untuk validasi data:"}
		case "2": // 2. Cek Progres Surat
			if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_CEK_PROGRES); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Silakan masukkan *Nomor Unik Surat* Anda (contoh: 1234):"}
		default:
			return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2."}
		}

	case STEP_SURAT_VALIDASI_NIK: // (Langkah 501)
		dataPenduduk, err := db.GetDataPendudukByNIK(dbConn, text)
		if err != nil {
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil { /*...*/ }
			return []string{"❌ NIK Anda tidak terdaftar di sistem kami.\n\n📝 Silakan pilih *Menu 1 (Data Diri)* untuk melakukan input data diri terlebih dahulu sebelum membuat surat.\n\n" + getMainMenu()}
		}
		
		// NIK Ditemukan. Simpan NIK ke kolom sesi DB
		if err := db.UpdateSessionField(dbConn, jid, "surat_valid_nik", text); err != nil {
             return []string{"Kesalahan menyimpan NIK. Coba lagi."}
        }
		
		// Auto-fill data dasar
		dataMap := surat.BuildDataMap(db.DataPenduduk(*dataPenduduk))
		dataMapBytes, _ := json.Marshal(dataMap)
		if err := db.UpdateSessionField(dbConn, jid, "surat_data_map", string(dataMapBytes)); err != nil {
			return []string{"Kesalahan menyimpan data map. Coba lagi."}
		}

		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_PILIH_JENIS); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{
			"NIK Tervalidasi. Silakan pilih jenis surat:\n" +
				"1. Surat Domisili\n2. Surat Usaha\n3. SKTM Umum\n4. SKTM Tanggungan\n5. Surat Kematian",
		}

	case STEP_SURAT_PILIH_JENIS: // (Langkah 502)
		jenisSurat, ok := surat.JenisSuratMap[text]
		if !ok {
			return []string{"❌ Pilihan tidak valid. Masukkan angka 1-5 sesuai jenis surat."}
		}
		
		if err := db.SetEditField(dbConn, jid, string(jenisSurat)); err != nil { /*...*/ } // Simpan jenis surat

		// Ambil NIK yg tersimpan
		nik, _ := db.GetSessionField(dbConn, jid, "surat_valid_nik")
		dataPenduduk, _ := db.GetDataPendudukByNIK(dbConn, nik.String)
		
		fieldList := surat.GetFieldList(db.DataPenduduk(*dataPenduduk), jenisSurat)

		if len(fieldList) == 0 { // Tidak ada field tambahan
			return handleSuratGeneration(dbConn, jid, session, sheetsClient, waClient)
		}
		
		fieldStr := strings.Join(fieldList, ",")
		if err := db.UpdateSessionField(dbConn, jid, "surat_fields_pending", fieldStr); err != nil { /*...*/ }
		if err := db.UpdateSessionField(dbConn, jid, "surat_field_now", fieldList[0]); err != nil { /*...*/ }
		
		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_INPUT_DATA); err != nil {
			return []string{"❌ Maaf, terjadi kesalahan sistem."}
		}
		
		return []string{surat.GetPrompt(fieldList[0])}
	
	case STEP_SURAT_INPUT_DATA: // (Langkah 503)
		// Simpan jawaban sementara (kita akan gunakan 'edit_field' sebagai 'temp_answer')
	if err := db.UpdateSessionField(dbConn, jid, "surat_temp_answer", text); err != nil {
    return []string{"Kesalahan menyimpan jawaban sementara."}
	}
		
		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_KONFIRMASI_FIELD); err != nil { /*...*/ }
		return []string{fmt.Sprintf("Anda mengisi: *%s*\n\nKetik 'ya' untuk lanjut, atau 'edit' untuk mengulangi.", text)}

	case STEP_SURAT_KONFIRMASI_FIELD: // (Langkah 504)
		currentField := session.SuratFieldNow.String
		fieldList := strings.Split(session.SuratFieldsPending.String, ",")
		
		if strings.ToLower(text) == "edit" {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_INPUT_DATA); err != nil { /*...*/ }
			return []string{surat.GetPrompt(currentField)} // Tanyakan lagi
		}

		if strings.ToLower(text) == "ya" {
			// Jawaban "ya", simpan data
			tempAnswer := session.SuratTempAnswer.String // (Kita sudah punya 'session' dari parameter)
    	dataMap := make(map[string]string)
			_ = json.Unmarshal([]byte(session.SuratDataMap.String), &dataMap)
			dataMap[currentField] = tempAnswer
			
			next := surat.NextField(fieldList, currentField)
			
			if next != "" { // Masih ada pertanyaan
				dataMapBytes, _ := json.Marshal(dataMap)
				if err := db.UpdateSessionField(dbConn, jid, "surat_data_map", string(dataMapBytes)); err != nil { /*...*/ }
				if err := db.UpdateSessionField(dbConn, jid, "surat_field_now", next); err != nil { /*...*/ }
				if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_INPUT_DATA); err != nil { /*...*/ }
				return []string{surat.GetPrompt(next)}
			}
			
			// Selesai, buat surat
			dataMapBytes, _ := json.Marshal(dataMap)
			if err := db.UpdateSessionField(dbConn, jid, "surat_data_map", string(dataMapBytes)); err != nil { /*...*/ }
			return handleSuratGeneration(dbConn, jid, session, sheetsClient, waClient)
		}
		
		return []string{"Pilihan tidak valid. Ketik 'ya' atau 'edit'."}

	case STEP_SURAT_CEK_PROGRES: // (Langkah 505)
		unikID := strings.ToUpper(text)
		status, err := sheetsClient.GetSuratStatus(unikID)
		if err != nil {
			return []string{fmt.Sprintf("Nomor Unik Surat *%s* tidak ditemukan.", unikID)}
		}
		if err := db.DeleteDataEntrySession(dbConn, jid); err != nil { /*...*/ }
		return []string{fmt.Sprintf("Status untuk surat *%s*:\n\n*STATUS: %s*", unikID, status)}

	// 3. alur fitur pengaduan
	case STEP_PENGADUAN_MENU:
		switch text {
		case "1": // 1. Ajukan Pengaduan
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_VALIDASI_NIK); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"📋 Untuk mengajukan pengaduan, silakan masukkan *NIK 16 digit* Anda untuk verifikasi:"}
		case "2": // 2. Cek Status Pengaduan
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_CARI_ID); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"🔍 Silakan masukkan *ID Pengaduan* Anda (contoh: P-101):"}
		default:
			return []string{"❌ Pilihan tidak valid. Silakan pilih 1 atau 2."}
		}

	case STEP_PENGADUAN_VALIDASI_NIK:
		// 1. Validasi format NIK (opsional tapi disarankan)
		if len(text) != 16 {
			return []string{"❌ Format NIK salah. Harap masukkan 16 digit NIK Anda:"}
		}

		// 2. Cek NIK ke database permanen
		isRegistered, err := db.CheckNIKExistsInPenduduk(dbConn, text)
		if err != nil {
			log.Printf("[ERROR] Gagal mengecek NIK di DB permanen: %v", err)
			return []string{"❌ Maaf, terjadi kesalahan sistem saat validasi NIK."}
		}

		if isRegistered {
			// 3. JIKA BERHASIL: Lanjutkan ke langkah kirim foto
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_WAITING); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"✅ NIK Anda terverifikasi. 📸 Silakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."}
		} else {
			// 4. JIKA GAGAL: Kembalikan ke menu utama
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil { // Hapus sesi
				log.Printf("[ERROR] Gagal hapus sesi setelah NIK tidak ditemukan: %v", err)
			}
			return []string{"❌ NIK Anda tidak terdaftar di sistem kami.\n\n📝 Silakan pilih *Menu 1 (Data Diri)* untuk melakukan input data diri terlebih dahulu sebelum membuat pengaduan.\n\n" + getMainMenu()}
		}

	case STEP_PENGADUAN_WAITING:
		// Jika user mengirim teks padahal kita menunggu gambar
		return []string{"📸 Mohon kirimkan *gambar* pengaduan, bukan teks. Atau ketik 'reset' untuk batal."}

	case STEP_PENGADUAN_CARI_ID:
		// User mengirimkan ID, kita cari di Sheet
		publicID := strings.ToUpper(text) // P-101

		// Panggil fungsi baru (akan kita buat di Fase 3)
		status, err := sheetsClient.GetPengaduanStatus(publicID)
		if err != nil {
			log.Printf("[WARN] Gagal mencari status untuk ID %s: %v", publicID, err)
			return []string{fmt.Sprintf("❌ ID Pengaduan *%s* tidak ditemukan. Pastikan Anda memasukkan ID dengan benar.", publicID)}
		}

		// Sukses! Hapus sesi dan kirim status
		if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Gagal hapus sesi setelah cek status: %v", err)
		}
		return []string{fmt.Sprintf("📋 Status untuk pengaduan *%s*:\n\n*STATUS: %s*", publicID, status)}
	// --- 1_2. ALUR LANGKAH VIRTUAL (dari fitur input) ---
	case STEP_NIK_DUPLICATE: // (101)
		text = strings.ToLower(text)
		if text == "edit nik" {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{formatQuestion(steps[STEP_NIK])}
		} else if text == "stop" {
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"❌ Pendaftaran dibatalkan.\n\n" + getMainMenu()}
		} else {
			return []string{"⚠️ Pilihan tidak valid.\n\n📝 Ketik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
		}

	case STEP_CHECKPOINT_DATA_INTI: // (100)
		text = strings.ToLower(text)
		if text == "lanjut" {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_GOLONGAN_DARAH); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{formatQuestion(steps[STEP_GOLONGAN_DARAH])}
		} else if text == "cukup" {
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"📝 Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		} else {
			return []string{"❌ Mohon ketik 'lanjut' atau 'cukup'."}
		}

	// --- 5. ALUR KONFIRMASI & EDIT (dari fitur input & edit) ---
	case STEP_CONFIRMATION, STEP_EDIT: // (42, 43)
		if session.CurrentStep == STEP_CONFIRMATION {
			text = strings.ToLower(text)
			if text != "valid" && text != "edit" {
				data, err := db.GetFormattedSessionData(dbConn, jid)
				if err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem."}
				}
				return []string{"📝 Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
			}
		}

		switch strings.ToLower(text) {
		case "valid":
			fullSession, err := db.GetFullSessionData(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Gagal mengambil data lengkap: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem saat mengambil data."}
			}

			// 1. Simpan ke Database (PostgreSQL) - Ini akan INSERT atau UPDATE
			if err := db.SaveDataPenduduk(dbConn, *fullSession); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan besar saat menyimpan ke database."}
			}

			// 2. Simpan ke Google Sheet (di background)
			go func() {
				nik := fullSession.NIK.String
				rowNum, err := sheetsClient.FindRowByNIK(nik)
				if err != nil {
					// NIK tidak ditemukan, berarti INPUT BARU
					log.Printf("Menambah NIK %s baru ke Sheet.", nik)
					sheetsClient.AppendDataPenduduk(*fullSession)
				} else {
					// NIK ditemukan, berarti EDIT
					log.Printf("Mengupdate NIK %s di baris %d Sheet.", nik, rowNum)
					sheetsClient.UpdateRowData(rowNum, *fullSession)
				}
			}()

			// 3. Hapus sesi sementara
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem"}
			}
			return []string{"✅ Terima kasih! Data Anda telah berhasil disimpan di Database dan sedang disinkronkan ke Spreadsheet."}

		case "edit":
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Failed to get formatted data: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem"}
			}
			if err := db.SetEditField(dbConn, jid, ""); err != nil {
				log.Printf("[ERROR] Failed to clear edit field: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem"}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT); err != nil {
				log.Printf("[ERROR] Failed to update step: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem"}
			}
			return []string{"📝 Ketik nomor yang ingin anda edit (1-39)\n\n" + data}

		default: // Ini adalah logika saat user mengedit field
			editField, err := db.GetEditField(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Failed to get edit field: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem"}
			}

			log.Printf("[DEBUG] STEP_EDIT - editField: '%s', isEmpty: %v", editField, editField == "")

			if session.CurrentStep == STEP_EDIT && editField == "" {
				// User belum memilih field untuk diedit
				num, err := strconv.Atoi(text)
				if err != nil || num < 1 || num > 39 {
					data, err := db.GetFormattedSessionData(dbConn, jid)
					if err != nil {
						return []string{"❌ Maaf, terjadi kesalahan sistem"}
					}
					return []string{fmt.Sprintf("⚠️ Nomor tidak valid\n\n📝 Silakan ketik:\n- Nomor 1-39 untuk mengedit data\n- 'valid' untuk menyimpan\n\n%s", data)}
				}

				step := steps[num]
				if err := db.SetEditField(dbConn, jid, step.Field); err != nil {
					log.Printf("[ERROR] Failed to set edit field: %v", err)
					return []string{"❌ Maaf, terjadi kesalahan sistem"}
				}
				if step.Options != nil {
					return []string{(formatQuestionWithOptionsForError(step.Question, step.Options))}
				}
				// For non-option fields, gunakan formatQuestion untuk menampilkan hint format
				return []string{(formatQuestion(step))}
			}

			if editField != "" {
				// User sedang mengisi nilai untuk field yang dipilih
				log.Printf("[DEBUG] Processing field edit for field: %s with input: %s", editField, text)
				
				var currentStep Step
				for _, step := range steps {
					if step.Field == editField {
						currentStep = step
						break
					}
				}
				
				// Jika step tidak ditemukan, kembalikan error
				if currentStep.Field == "" {
					log.Printf("[ERROR] Step not found for field: %s", editField)
					return []string{"❌ Maaf, terjadi kesalahan sistem - field tidak ditemukan"}
				}
				
				value, err := validateInput(text, currentStep)
				if err != nil {
					return []string{err.Error()}
				}
				if err := db.UpdateDataEntrySession(dbConn, jid, editField, value); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem"}
				}
				if err := db.SetEditField(dbConn, jid, ""); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem"}
				}
				data, err := db.GetFormattedSessionData(dbConn, jid)
				if err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem"}
				}
				if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem"}
				}
				return []string{
					"✅ Data berhasil diupdate\n\n📝 Silakan ketik:\n- Nomor 1-39 untuk mengedit data lainnya\n- 'valid' jika sudah selesai\n\n" + data,
				}
			}

			// Jika sampai sini berarti ada kondisi tidak terduga
			log.Printf("[ERROR] Unexpected state in STEP_EDIT - editField: '%s'", editField)
			return []string{"❌ Maaf, terjadi kesalahan sistem - kondisi tidak valid"}
		}

	// --- 6. ALUR NORMAL INPUT DATA (LANGKAH 1-41) ---
	default:
		stepInfo, ok := steps[session.CurrentStep]
		if !ok { // Seharusnya ini adalah langkah setelah 39
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"📝 Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		}

		// (Khusus untuk langkah 1)
		if session.CurrentStep == STEP_START && text == "1" {
			return []string{formatQuestion(steps[STEP_START])}
		}

		// Validasi input
		value, err := validateInput(text, stepInfo)
		if err != nil {
			log.Printf("[DEBUG] Input validation failed: %v", err)
			return []string{err.Error()}
		}

		// --- VALIDASI NIK (saat input baru) ---
		if session.CurrentStep == STEP_NIK {
			nikValue := value.(string)
			// 1. Cek ke DB permanen (Cepat)
			isDuplicate, err := db.CheckNIKExistsInPenduduk(dbConn, nikValue)
			if err != nil {
				log.Printf("[ERROR] Gagal mengecek NIK di DB permanen: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem saat validasi NIK (DB)."}
			}
			if isDuplicate {
				if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK_DUPLICATE); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem."}
				}
				return []string{"⚠️ NIK ini sudah terdaftar di data penduduk desa.\n\n📝 Ketik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
			}
			// 2. Cek ke Sesi lain (Sementara)
			isDuplicate, err = db.CheckNIKExists(dbConn, nikValue, jid)
			if err != nil {
				log.Printf("[ERROR] Gagal mengecek NIK di DB sesi: %v", err)
				return []string{"❌ Maaf, terjadi kesalahan sistem saat validasi NIK (Sesi)."}
			}
			if isDuplicate {
				if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK_DUPLICATE); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem."}
				}
				return []string{"⚠️ NIK ini sedang didaftarkan oleh pengguna lain saat ini.\n\n📝 Ketik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
			}
		}
		// --- AKHIR VALIDASI NIK ---

		// Simpan data ke sesi sementara
		if err := db.UpdateDataEntrySession(dbConn, jid, stepInfo.Field, value); err != nil {
			log.Printf("[ERROR] Failed to update session: %v", err)
			if strings.Contains(err.Error(), "violates foreign key constraint") {
				return []string{fmt.Sprintf("⚠️ Pilihan tidak valid. Silakan pilih nomor dari daftar berikut:\n\n%s",
					formatQuestionWithOptions(stepInfo.Question, stepInfo.Options))}
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
		}

		// Cek checkpoint NAMA IBU
		if session.CurrentStep == STEP_DATA_INTI {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CHECKPOINT_DATA_INTI); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"✅ Data inti (sampai Suku) sudah tersimpan.\n\n❓ Apakah Anda ingin melanjutkan mengisi data pelengkap (NIK Ayah, NIK Ibu, Golongan Darah, dll)??\n\n📝 Ketik 'lanjut' untuk melanjutkan atau 'cukup' jika sudah selesai."}
		}

		// Jika sudah 39 langkah, tampilkan konfirmasi
		if session.CurrentStep == 39 { // STEP sebelum CONFIRMATION
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
				return []string{"❌ Maaf, terjadi kesalahan sistem."}
			}
			return []string{"📝 Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		}

		// Untuk step normal (1-38), lanjut ke step berikutnya dengan pertanyaan berikutnya
		if session.CurrentStep < 39 {
			nextStep := session.CurrentStep + 1
			if nextStepInfo, ok := steps[nextStep]; ok {
				if err := db.UpdateStepOnly(dbConn, jid, nextStep); err != nil {
					return []string{"❌ Maaf, terjadi kesalahan sistem."}
				}
				return []string{formatQuestion(nextStepInfo)}
			}
		}

		return []string{"❌ Maaf, terjadi kesalahan sistem."}
	}

	// Default catch-all (seharusnya tidak pernah tercapai)
	return []string{"❌ Terjadi error pada alur. Sesi direset.\n\n" + getMainMenu()}
}

func handleSuratGeneration(dbConn *sqlx.DB, jid string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client) []string {
	// Ambil sesi terbaru (termasuk data map yang baru diupdate)
	fullSession, err := db.GetFullSessionData(dbConn, jid)
	if err != nil {
		log.Printf("[SURAT-ERROR] Gagal mengambil data sesi lengkap: %v", err)
		return []string{"Terjadi kesalahan saat mengambil data sesi Anda."}
	}

	dataMap := make(map[string]string)
	if err := json.Unmarshal([]byte(fullSession.SuratDataMap.String), &dataMap); err != nil {
		log.Printf("[SURAT-ERROR] Gagal unmarshal data map: %v", err)
		return []string{"Terjadi kesalahan data map. Silakan 'reset'."}
	}
	
	jenisStr := fullSession.EditField.String
	jenis := surat.JenisSurat(jenisStr)
	
	// 1. Buat ID unik & nama file
	unikID := fmt.Sprintf("%04d", 1000+rand.Intn(9000)) // 4 digit random
	tgl := time.Now().Format("02-01-2006")
	// Format: sk_domisili_1234_02-01-2006.pdf
	pdfName := fmt.Sprintf("%s_%s_%s.pdf", strings.TrimSuffix(string(jenis), ".tex"), unikID, tgl)

	// 2. Upload PDF ke Cloud (ImgBB)
	// (Kita lakukan ini DULU agar bisa kirim link ke Kades jika PDF gagal dikirim)
	// (Implementasi: Anda perlu menjalankan Generate, LALU baca file PDF, LALU upload)

	// 3. Panggil generator (akan mengirim ke user + kades)
	_, err = surat.GenerateAsync(jenis, dataMap, "temp", jid, waClient, pdfName, unikID, sheetsClient)
	if err != nil {
		log.Printf("[SURAT-ERROR] %v", err)
		return []string{"Terjadi kesalahan saat memproses surat."}
	}


	// 5. Hapus sesi
	db.DeleteDataEntrySession(dbConn, jid)

	return []string{
		fmt.Sprintf("Surat *%s* Anda sedang diproses dan akan segera dikirimkan. Harap tunggu...", surat.NamaSuratmap[jenis]),
	}
}

// Helper function untuk mendapatkan label field yang user-friendly
func getFieldLabel(field string) string {
	fieldLabels := map[string]string{
		"nik":                    "NIK",
		"no_kk":                  "No. KK",
		"nik_ayah":               "NIK Ayah",
		"nik_ibu":                "NIK Ibu",
		"dusun":                  "Dusun",
		"rt":                     "RT",
		"nama":                   "Nama",
		"nama_ayah":              "Nama Ayah",
		"nama_ibu":               "Nama Ibu",
		"tempat_lahir":           "Tempat Lahir",
		"alamat_sekarang":        "Alamat Sekarang",
		"tag_card":               "Tag Card",
		"no_asuransi":            "No. Asuransi",
		"akta_lahir":             "No. Akta Lahir",
		"dokumen_passport":       "No. Dokumen Paspor",
		"dokumen_kitas":          "No. Dokumen KITAS",
		"akta_perkawinan":        "No. Akta Perkawinan",
		"akta_perceraian":        "No. Akta Perceraian",
	}
	if label, ok := fieldLabels[field]; ok {
		return label
	}
	return strings.ToLower(field)
}

// Helper function untuk format pertanyaan dengan options TANPA hint di awal (untuk error)
func formatQuestionWithOptionsForError(question string, options map[int]string) string {
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
	
	// Add example at the end with spasi
	builder.WriteString("\n\n📝 Contoh input: 1")
	
	return builder.String()
}

func validateInput(text string, step Step) (interface{}, error) {
	text = strings.TrimSpace(text)
	label := getFieldLabel(step.Field)

	// Special validation for empty input
	if text == "" {
		return nil, fmt.Errorf("input tidak boleh kosong\n\n%s", step.Question)
	}
	if step.IsDate {
		t, err := time.Parse("02-01-2006", text)
		if err != nil {
			return nil, fmt.Errorf("⚠ Format tanggal salah. Harap masukkan dengan format DD-MM-YYYY (contoh: 25-12-2024)\n\n%s", step.Question)
		}
		return t, nil
	}
	// Handle options validation first
	if step.Options != nil {
		choice, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("⚠️ Input harus berupa angka. Silakan pilih nomor dari daftar berikut:\n\n%s",
				formatQuestionWithOptionsForError(step.Question, step.Options))
		}

		// Validate against available options
		if _, valid := step.Options[choice]; !valid {
			return nil, fmt.Errorf("⚠️ Pilihan nomor %d tidak tersedia. Silakan pilih nomor dari daftar berikut:\n\n%s",
				choice, formatQuestionWithOptionsForError(step.Question, step.Options))
		}

		return choice, nil
	}

	// Helper function to build error message with example
	buildErrorMsg := func(errorMsg string) string {
		var result strings.Builder
		result.WriteString(errorMsg)
		result.WriteString("\n\n")
		result.WriteString(step.Question)

		// Add example based on field type
		switch step.Field {
		case "nik", "no_kk", "nik_ayah", "nik_ibu":
			result.WriteString("\n\n📝 Contoh: 3173012345678901")
		case "rt", "dusun":
			result.WriteString("\n\n📝 Contoh: 002")
		case "nama", "nama_ayah", "nama_ibu":
			if step.Field == "nama_ayah" {
				result.WriteString("\n\n📝 Contoh: Ahmad Suhaimi")
			} else {
				result.WriteString("\n\n📝 Contoh: Ida Suheti")
			}
		case "tempat_lahir":
			result.WriteString("\n\n📝 Contoh: Bandung")
		case "tanggal_lahir":
			result.WriteString("\n\n📝 Contoh: 15-05-1990")
		case "alamat_sekarang":
			result.WriteString("\n\n📝 Contoh: Jl. Merdeka No. 123")
		case "akta_lahir", "dokumen_passport", "dokumen_kitas", "akta_perkawinan", "akta_perceraian":
			result.WriteString("\n\n📝 Contoh: AK-2023-000123")
		case "tanggal_akhir_passport", "tanggal_perkawinan", "tanggal_perceraian":
			result.WriteString("\n\n📝 Contoh: 31-12-2030")
		case "tag_card":
			result.WriteString("\n\n📝 Contoh: TAG-2023-001")
		case "no_asuransi":
			result.WriteString("\n\n📝 Contoh: JKN-1234567890")
		}
		return result.String()
	}

	// Field-specific validations
	switch step.Field {
	case "nik", "no_kk", "nik_ayah", "nik_ibu":
		if len(text) != 16 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s harus 16 digit, (saat ini: %d)", label, len(text))))
		}
		if _, err := strconv.ParseInt(text, 10, 64); err != nil {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s hanya boleh berisi angka", label)))
		}
		return text, nil
	case "dusun", "rt":
		if len(text) != 3 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s harus 3 digit, (saat ini: %d)", label, len(text))))
		}
		if _, err := strconv.Atoi(text); err != nil {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s hanya boleh berisi angka", label)))
		}
		return text, nil
	case "nama", "nama_ayah", "nama_ibu":
		if len(text) < 2 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s terlalu pendek, minimal 2 karakter (saat ini: %d)", label, len(text))))
		}
		if strings.ContainsAny(text, "0123456789") {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s tidak boleh mengandung angka", label)))
		}
		if strings.ContainsAny(text, "!@#$%^&*()_+=[]{};:'\"\\|,.<>/?") {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s tidak boleh mengandung karakter khusus", label)))
		}
	case "tempat_lahir":
		if len(text) < 3 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s terlalu pendek, minimal 3 karakter (saat ini: %d)", label, len(text))))
		}
		if strings.ContainsAny(text, "0123456789") {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s tidak boleh mengandung angka", label)))
		}
	case "alamat_sekarang":
		if len(text) < 5 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s terlalu pendek, minimal 5 karakter (saat ini: %d)", label, len(text))))
		}
	case "tag_card":
		if len(text) < 3 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s terlalu pendek, minimal 3 karakter (saat ini: %d)", label, len(text))))
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9-/]+$`).MatchString(text) {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s hanya boleh berisi huruf, angka, dash (-) dan garis miring (/)", label)))
		}
	case "no_asuransi":
		if len(text) < 3 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s terlalu pendek, minimal 3 karakter (saat ini: %d)", label, len(text))))
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9-/]+$`).MatchString(text) {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ %s hanya boleh berisi huruf, angka, dash (-) dan garis miring (/)", label)))
		}
	case "akta_lahir", "dokumen_passport", "dokumen_kitas", "akta_perkawinan", "akta_perceraian":
		if len(text) < 3 {
			return nil, fmt.Errorf(buildErrorMsg(fmt.Sprintf("⚠️ nomor dokumen terlalu pendek, minimal 3 karakter (saat ini: %d)", len(text))))
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9-/]+$`).MatchString(text) {
			return nil, fmt.Errorf(buildErrorMsg("⚠️ nomor dokumen hanya boleh berisi huruf, angka, dash (-) dan garis miring (/)"))
		}
	}

	return text, nil
}

func formatQuestion(step Step) string {
	if step.Options != nil {
		return formatQuestionWithOptions(step.Question, step.Options)
	}

	var result strings.Builder
	
	// Add format hint at the beginning for all questions
	result.WriteString("💡 Mohon masukkan data sesuai format yang tersedia.\n\n")
	result.WriteString(step.Question)

	// Add hints for specific fields
	switch step.Field {
	case "nik", "no_kk", "nik_ayah", "nik_ibu":
		result.WriteString("\n\n📝 Contoh: 3173012345678901")
	case "rt", "dusun":
		result.WriteString("\n\n📝 Contoh: 002")
	case "nama", "nama_ayah", "nama_ibu":
		if step.Field == "nama_ayah" {
			result.WriteString("\n\n📝 Contoh: Ahmad Suhaimi")
		} else {
			result.WriteString("\n\n📝 Contoh: Ida Suheti")
		}
	case "tempat_lahir":
		result.WriteString("\n\n📝 Contoh: Bandung")
	case "tanggal_lahir":
		result.WriteString("\n\n📝 Contoh: 15-05-1990")
	case "alamat_sekarang":
		result.WriteString("\n\n📝 Contoh: Jl. Merdeka No. 123")
	case "akta_lahir", "dokumen_passport", "dokumen_kitas", "akta_perkawinan", "akta_perceraian":
		result.WriteString("\n\n📝 Contoh: AK-2023-000123")
	case "tanggal_akhir_passport", "tanggal_perkawinan", "tanggal_perceraian":
		result.WriteString("\n\n📝 Contoh: 31-12-2030")
	case "tag_card":
		result.WriteString("\n\n📝 Contoh: TAG-2023-001")
	case "no_asuransi":
		result.WriteString("\n\n📝 Contoh: JKN-1234567890")
	}

	return result.String()
}

func formatQuestionWithOptions(question string, options map[int]string) string {
	var builder strings.Builder
	
	// Add format hint at the beginning
	builder.WriteString("💡 Mohon masukkan pilihan sesuai format yang tersedia.\n\n")
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
	
	// Add example at the end with spasi
	builder.WriteString("\n\n📝 Contoh input: 1")
	
	return builder.String()
}
