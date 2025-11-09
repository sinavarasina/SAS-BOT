// Gw pengen ini ntar di pecah jadi module baru di dalam folder bot/

package bot

import (
	"encoding/json"
	"fmt"
	"go.mau.fi/whatsmeow"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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

const (
	STEP_START                  = 1
	STEP_NIK                    = 7
	STEP_NAMA_IBU               = 21
	STEP_GOLONGAN_DARAH         = 22
	STEP_CONFIRMATION           = 42
	STEP_EDIT                   = 43
	STEP_CHECKPOINT_NAMA_IBU    = 100
	STEP_NIK_DUPLICATE          = 101
	STEP_MENU_DATA_DIRI         = 200 // Sesi menunggu pilihan (Input, Edit, Hapus)
	STEP_EDIT_CARI_NIK          = 201 // Sesi menunggu NIK untuk di-edit
	STEP_PENGADUAN_WAITING      = 301
	STEP_PENGADUAN_MENU         = 300
	STEP_PENGADUAN_CARI_ID      = 302
	STEP_PENGADUAN_VALIDASI_NIK = 303
	STEP_SURAT_MENU_UTAMA       = 500
	STEP_SURAT_VALIDASI_NIK     = 501
	STEP_SURAT_INPUT_DATA       = 502
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

func HandleDataEntry(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client) []string {
	log.Printf("[DEBUG] Handling data entry for step %d with input: '%s'", session.CurrentStep, text)

	// Handle "fast" command first
	if text == "fast" {
		if !session.AwaitingAnswer || session.CurrentStep < STEP_START {
			return []string{"Silakan pilih menu. Kirim '1' untuk memulai pendataan."}
		}
		if err := FastTrackDataEntry(dbConn, jid); err != nil {
			log.Printf("[ERROR] Fast track failed: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		data, err := db.GetFormattedSessionData(dbConn, jid)
		if err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	}

	switch session.CurrentStep {

	// --- 1. ALUR MENU UTAMA & SUB-MENU ---
	case STEP_MENU_DATA_DIRI: // User ada di sub-menu (200)
		switch text {
		case "1": // 1. Input Data Diri
			if err := db.StartNewSession(dbConn, jid); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{steps[STEP_START].Question}
		case "2": // 2. Edit Data Diri
			if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT_CARI_NIK); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Silakan masukkan **NIK 16 digit** yang datanya ingin Anda edit:"}
		default:
			return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2, atau ketik 'reset'."}
		}

	// --- 2. ALUR FITUR EDIT ---
	case STEP_EDIT_CARI_NIK: // Menunggu NIK untuk di-edit (201)
		data, err := db.GetDataPendudukByNIK(dbConn, text)
		if err != nil {
			log.Printf("[DEBUG] NIK %s tidak ditemukan di DB: %v", text, err)
			return []string{"NIK tidak ditemukan di database. Silakan coba lagi atau ketik 'reset'."}
		}
		// Temukan! Salin data permanen ke sesi sementara
		if err := db.LoadSessionFromPenduduk(dbConn, jid, *data); err != nil {
			log.Printf("[ERROR] Gagal LoadSessionFromPenduduk: %v", err)
			return []string{"NIK ditemukan, tapi gagal memuat data ke sesi. Hubungi admin."}
		}
		// Langsung lompat ke langkah konfirmasi (42)
		dataStr, _ := db.GetFormattedSessionData(dbConn, jid)
		return []string{"Data ditemukan. Silakan periksa:\n\n" + dataStr, "\n\nKetik 'valid' untuk menyimpan atau 'edit' untuk mengubah data."}

	// Menu 2
	case STEP_SURAT_MENU_UTAMA:
		jenisSurat, ok := surat.JenisSuratMap[text]
		if !ok {
			return []string{"Pilihan tidak valid. Masukkan angka 1-5 sesuai jenis surat."}
		}

		if err := db.SetEditField(dbConn, jid, string(jenisSurat)); err != nil {
			log.Printf("[SURAT_DB] Gagal menyimpan jenis surat: %v", err)
			return []string{"Maaf,terjadi kesalahan sistem."}
		}

		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_VALIDASI_NIK); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Baik, Anda memilih " + surat.NamaSuratmap[jenisSurat] + ".\n\n" + "Untuk melanjutkan, masukkan **NIK 16 Digit** anda untuk validasi data:"}

	case STEP_SURAT_VALIDASI_NIK:
		dataPenduduk, err := db.GetDataPendudukByNIK(dbConn, text)
		if err != nil {
			// NIK tidak ditemukan di DB permanen
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil { // Reset sesi
				log.Printf("[ERROR] Gagal hapus sesi setelah NIK surat tidak ditemukan: %v", err)
			}
			return []string{"NIK Anda tidak terdaftar di sistem kami.\n\n" +
				"Silakan pilih *Menu 1 (Data Diri)* untuk melakukan input data diri terlebih dahulu sebelum membuat surat.\n\n" +
				getMainMenu()}
		}

		// 2. Simpan NIK yang valid untuk sementara
		_ = db.SaveTemporary(fmt.Sprintf("surat_valid_nik_%s", jid), text)

		// 3. NIK Ditemukan! Siapkan alur input data surat
		jenisSuratStr, _ := db.GetEditField(dbConn, jid) // Ambil jenis surat yg disimpan
		fieldList := surat.GetFieldList(db.DataPenduduk(*dataPenduduk), surat.JenisSurat(jenisSuratStr))

		if len(fieldList) == 0 {
			// Jika surat tidak butuh input tambahan (langsung buat)
			return handleSuratGeneration(dbConn, jid, session, sheetsClient, waClient)
		}

		// 4. Simpan daftar field yang harus diisi
		fieldStr := strings.Join(fieldList, ",")
		_ = db.SaveTemporary(fmt.Sprintf("surat_fields_%s", jid), fieldStr)
		_ = db.SaveTemporary(fmt.Sprintf("surat_field_now_%s", jid), fieldList[0])

		if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_INPUT_DATA); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}

		// 5. Ajukan pertanyaan pertama
		return []string{surat.GetPrompt(fieldList[0])}

	case STEP_SURAT_INPUT_DATA:
		currentField, _ := db.LoadTemporary("surat_field_now_" + jid)
		fieldListStr, _ := db.LoadTemporary("surat_fields_" + jid)
		fieldList := strings.Split(fieldListStr, ",")

		_ = db.SaveTemporary(jid+"_field_"+currentField, text)

		next := surat.NextField(fieldList, currentField)
		if next != "" {
			// Masih ada pertanyaan
			_ = db.SaveTemporary("surat_field_now_"+jid, next)
			return []string{surat.GetPrompt(next)}
		}

		return handleSuratGeneration(dbConn, jid, session, sheetsClient, waClient)

	// 3. alur fitur pengaduan
	case STEP_PENGADUAN_MENU:
		switch text {
		case "1": // 1. Ajukan Pengaduan
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_VALIDASI_NIK); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Untuk mengajukan pengaduan, silakan masukkan **NIK 16 digit** Anda untuk verifikasi:"}
		case "2": // 2. Cek Status Pengaduan
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_CARI_ID); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Silakan masukkan *ID Pengaduan* Anda (contoh: P-101):"}
		default:
			return []string{"Pilihan tidak valid. Silakan pilih 1 atau 2."}
		}

	case STEP_PENGADUAN_VALIDASI_NIK:
		// 1. Validasi format NIK (opsional tapi disarankan)
		if len(text) != 16 {
			return []string{"Format NIK salah. Harap masukkan 16 digit NIK Anda:"}
		}

		// 2. Cek NIK ke database permanen
		isRegistered, err := db.CheckNIKExistsInPenduduk(dbConn, text)
		if err != nil {
			log.Printf("[ERROR] Gagal mengecek NIK di DB permanen: %v", err)
			return []string{"Maaf, terjadi kesalahan sistem saat validasi NIK."}
		}

		if isRegistered {
			// 3. JIKA BERHASIL: Lanjutkan ke langkah kirim foto
			if err := db.UpdateStepOnly(dbConn, jid, STEP_PENGADUAN_WAITING); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"NIK Anda terverifikasi. Silakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."}
		} else {
			// 4. JIKA GAGAL: Kembalikan ke menu utama
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil { // Hapus sesi
				log.Printf("[ERROR] Gagal hapus sesi setelah NIK tidak ditemukan: %v", err)
			}
			return []string{"NIK Anda tidak terdaftar di sistem kami.\n\nSilakan pilih *Menu 1 (Data Diri)* untuk melakukan input data diri terlebih dahulu sebelum membuat pengaduan.\n\n" + getMainMenu()}
		}

	case STEP_PENGADUAN_WAITING:
		// Jika user mengirim teks padahal kita menunggu gambar
		return []string{"Mohon kirimkan *gambar* pengaduan, bukan teks. Atau ketik 'reset' untuk batal."}

	case STEP_PENGADUAN_CARI_ID:
		// User mengirimkan ID, kita cari di Sheet
		publicID := strings.ToUpper(text) // P-101

		// Panggil fungsi baru (akan kita buat di Fase 3)
		status, err := sheetsClient.GetPengaduanStatus(publicID)
		if err != nil {
			log.Printf("[WARN] Gagal mencari status untuk ID %s: %v", publicID, err)
			return []string{fmt.Sprintf("ID Pengaduan *%s* tidak ditemukan. Pastikan Anda memasukkan ID dengan benar.", publicID)}
		}

		// Sukses! Hapus sesi dan kirim status
		if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
			log.Printf("[ERROR] Gagal hapus sesi setelah cek status: %v", err)
		}
		return []string{fmt.Sprintf("Status untuk pengaduan *%s*:\n\n*STATUS: %s*", publicID, status)}
	// --- 1_2. ALUR LANGKAH VIRTUAL (dari fitur input) ---
	case STEP_NIK_DUPLICATE: // (101)
		text = strings.ToLower(text)
		if text == "edit nik" {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{formatQuestion(steps[STEP_NIK])}
		} else if text == "stop" {
			if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Pendaftaran dibatalkan.\n\n" + getMainMenu()}
		} else {
			return []string{"⚠️ Pilihan tidak valid.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
		}

	case STEP_CHECKPOINT_NAMA_IBU: // (100)
		text = strings.ToLower(text)
		if text == "lanjut" {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_GOLONGAN_DARAH); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{formatQuestion(steps[STEP_GOLONGAN_DARAH])}
		} else if text == "cukup" {
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		} else {
			return []string{"Mohon ketik 'lanjut' atau 'cukup'."}
		}

	// --- 5. ALUR KONFIRMASI & EDIT (dari fitur input & edit) ---
	case STEP_CONFIRMATION, STEP_EDIT: // (42, 43)
		if session.CurrentStep == STEP_CONFIRMATION {
			text = strings.ToLower(text)
			if text != "valid" && text != "edit" {
				data, err := db.GetFormattedSessionData(dbConn, jid)
				if err != nil {
					return []string{"Maaf, terjadi kesalahan sistem."}
				}
				return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
			}
		}

		switch strings.ToLower(text) {
		case "valid":
			fullSession, err := db.GetFullSessionData(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Gagal mengambil data lengkap: %v", err)
				return []string{"Maaf, terjadi kesalahan sistem saat mengambil data."}
			}

			// 1. Simpan ke Database (PostgreSQL) - Ini akan INSERT atau UPDATE
			if err := db.SaveDataPenduduk(dbConn, *fullSession); err != nil {
				return []string{"Maaf, terjadi kesalahan besar saat menyimpan ke database."}
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
				return []string{"maaf, terjadi kesalahan sistem"}
			}
			return []string{"✅ terima kasih! Data Anda telah berhasil disimpan di Database dan sedang disinkronkan ke Spreadsheet."}

		case "edit":
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Failed to get formatted data: %v", err)
				return []string{"maaf, terjadi kesalahan sistem"}
			}
			if err := db.SetEditField(dbConn, jid, ""); err != nil {
				log.Printf("[ERROR] Failed to clear edit field: %v", err)
				return []string{"maaf, terjadi kesalahan sistem"}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT); err != nil {
				log.Printf("[ERROR] Failed to update step: %v", err)
				return []string{"maaf, terjadi kesalahan sistem"}
			}
			return []string{"Ketik nomor yang ingin anda edit (1-41)\n\n" + data}

		default: // Ini adalah logika saat user mengedit field
			editField, err := db.GetEditField(dbConn, jid)
			if err != nil {
				log.Printf("[ERROR] Failed to get edit field: %v", err)
				return []string{"maaf, terjadi kesalahan sistem"}
			}

			if session.CurrentStep == STEP_EDIT && editField == "" {
				if text == "valid" {
					// ... (Logika 'valid' duplikat)
					fullSession, err := db.GetFullSessionData(dbConn, jid)
					if err != nil {
						log.Printf("[ERROR] Gagal mengambil data lengkap: %v", err)
						return []string{"Maaf, terjadi kesalahan sistem saat mengambil data."}
					}
					if err := db.SaveDataPenduduk(dbConn, *fullSession); err != nil {
						return []string{"Maaf, terjadi kesalahan besar saat menyimpan ke database."}
					}
					go func() {
						nik := fullSession.NIK.String
						rowNum, err := sheetsClient.FindRowByNIK(nik)
						if err != nil {
							sheetsClient.AppendDataPenduduk(*fullSession)
						} else {
							sheetsClient.UpdateRowData(rowNum, *fullSession)
						}
					}()
					if err := db.DeleteDataEntrySession(dbConn, jid); err != nil {
						return []string{"maaf, terjadi kesalahan sistem"}
					}
					return []string{"✅ terima kasih! semua data telah berhasil dimasukkan"}
				}

				num, err := strconv.Atoi(text)
				if err != nil || num < 1 || num > 41 {
					data, err := db.GetFormattedSessionData(dbConn, jid)
					if err != nil {
						return []string{"maaf, terjadi kesalahan sistem"}
					}
					return []string{fmt.Sprintf("⚠️ nomor tidak valid\n\nsilakan ketik:\n- nomor 1-41 untuk mengedit data\n- 'valid' untuk menyimpan\n\n%s", data)}
				}

				step := steps[num]
				if err := db.SetEditField(dbConn, jid, step.Field); err != nil {
					log.Printf("[ERROR] Failed to set edit field: %v", err)
					return []string{"maaf, terjadi kesalahan sistem"}
				}
				if step.Options != nil {
					return []string{fmt.Sprintf("📝 edit data nomor %d:\n\n%s", num, formatQuestionWithOptions(step.Question, step.Options))}
				}
				return []string{fmt.Sprintf("📝 edit data nomor %d:\n\n%s", num, formatQuestion(step))}
			}

			if editField != "" {
				var currentStep Step
				for _, step := range steps {
					if step.Field == editField {
						currentStep = step
						break
					}
				}
				value, err := validateInput(text, currentStep)
				if err != nil {
					return []string{fmt.Sprintf("📝 mode edit aktif\n\n%s", err.Error())}
				}
				if err := db.UpdateDataEntrySession(dbConn, jid, editField, value); err != nil {
					return []string{"maaf, terjadi kesalahan sistem"}
				}
				if err := db.SetEditField(dbConn, jid, ""); err != nil {
					return []string{"maaf, terjadi kesalahan sistem"}
				}
				data, err := db.GetFormattedSessionData(dbConn, jid)
				if err != nil {
					return []string{"maaf, terjadi kesalahan sistem"}
				}
				if err := db.UpdateStepOnly(dbConn, jid, STEP_EDIT); err != nil {
					return []string{"maaf, terjadi kesalahan sistem"}
				}
				return []string{
					"✅ data berhasil diupdate\n\nketik:\n- nomor 1-41 untuk mengedit data lain\n- 'valid' jika sudah selesai\n\n" + data,
				}
			}
		}

	// --- 6. ALUR NORMAL INPUT DATA (LANGKAH 1-41) ---
	default:
		stepInfo, ok := steps[session.CurrentStep]
		if !ok { // Seharusnya ini adalah langkah setelah 41
			data, err := db.GetFormattedSessionData(dbConn, jid)
			if err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
		}

		// (Khusus untuk langkah 1)
		if session.CurrentStep == STEP_START && text == "1" {
			return []string{steps[STEP_START].Question}
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
				return []string{"Maaf, terjadi kesalahan sistem saat validasi NIK (DB)."}
			}
			if isDuplicate {
				if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK_DUPLICATE); err != nil {
					return []string{"Maaf, terjadi kesalahan sistem."}
				}
				return []string{"⚠️ NIK ini sudah terdaftar di data penduduk desa.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
			}
			// 2. Cek ke Sesi lain (Sementara)
			isDuplicate, err = db.CheckNIKExists(dbConn, nikValue, jid)
			if err != nil {
				log.Printf("[ERROR] Gagal mengecek NIK di DB sesi: %v", err)
				return []string{"Maaf, terjadi kesalahan sistem saat validasi NIK (Sesi)."}
			}
			if isDuplicate {
				if err := db.UpdateStepOnly(dbConn, jid, STEP_NIK_DUPLICATE); err != nil {
					return []string{"Maaf, terjadi kesalahan sistem."}
				}
				return []string{"⚠️ NIK ini sedang didaftarkan oleh pengguna lain saat ini.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran."}
			}
		}
		// --- AKHIR VALIDASI NIK ---

		// Simpan data ke sesi sementara
		if err := db.UpdateDataEntrySession(dbConn, jid, stepInfo.Field, value); err != nil {
			log.Printf("[ERROR] Failed to update session: %v", err)
			if strings.Contains(err.Error(), "violates foreign key constraint") {
				return []string{fmt.Sprintf("⚠️ pilihan tidak valid\n\n%s",
					formatQuestionWithOptions(stepInfo.Question, stepInfo.Options))}
			}
			return []string{"Maaf, terjadi kesalahan sistem."}
		}

		// Cek checkpoint NAMA IBU
		if session.CurrentStep == STEP_NAMA_IBU {
			if err := db.UpdateStepOnly(dbConn, jid, STEP_CHECKPOINT_NAMA_IBU); err != nil {
				return []string{"Maaf, terjadi kesalahan sistem."}
			}
			return []string{"Data inti (sampai Nama Ibu) sudah tersimpan.\n\n" +
				"Apakah Anda ingin melanjutkan mengisi data pelengkap (Gol. Darah, Akta, Asuransi, dll)?\n\n" +
				"Ketik 'lanjut' atau 'cukup'."}
		}

		// Lanjut ke pertanyaan berikutnya
		nextStep := session.CurrentStep + 1
		if stepInfo, ok := steps[nextStep]; ok {
			return []string{formatQuestion(stepInfo)}
		}

		// Jika sudah 41 langkah, tampilkan konfirmasi
		data, err := db.GetFormattedSessionData(dbConn, jid)
		if err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		if err := db.UpdateStepOnly(dbConn, jid, STEP_CONFIRMATION); err != nil {
			return []string{"Maaf, terjadi kesalahan sistem."}
		}
		return []string{"Ketik 'valid' jika sudah benar atau ketik 'edit' untuk mengubah data.\n\n" + data}
	}

	// Default catch-all (seharusnya tidak pernah tercapai)
	return []string{"Terjadi error pada alur. Sesi direset.\n\n" + getMainMenu()}
}

func handleSuratGeneration(dbConn *sqlx.DB, jid string, session *db.DataEntrySession, sheetsClient *sheets.SheetsClient, waClient *whatsmeow.Client) []string {

	// 1. Ambil NIK yang sudah divalidasi
	nik, ok := db.LoadTemporary(fmt.Sprintf("surat_valid_nik_%s", jid))
	if !ok {
		log.Printf("[SURAT-ERROR] Gagal mengambil NIK tervalidasi dari sesi JID %s", jid)
		return []string{"Terjadi kesalahan sesi (NIK tidak ditemukan). Silakan 'reset' dan coba lagi."}
	}

	// 2. Ambil data lengkap penduduk dari DB
	dataPenduduk, err := db.GetDataPendudukByNIK(dbConn, nik)
	if err != nil {
		log.Printf("[SURAT-ERROR] Gagal mengambil data penduduk (NIK: %s): %v", nik, err)
		return []string{"Terjadi kesalahan saat mengambil data penduduk Anda."}
	}

	// 3. Bangun data dasar (Nama, NIK, Alamat, dll)
	data := surat.BuildDataMap(db.DataPenduduk(*dataPenduduk))

	// 4. Ambil data tambahan yang baru diinput (misal: ALASANPERLU)
	fieldListStr, _ := db.LoadTemporary("surat_fields_" + jid)
	for f := range strings.SplitSeq(fieldListStr, ",") {
		f = strings.TrimSpace(f)
		val, _ := db.LoadTemporary(jid + "_field_" + f)
		data[f] = val
	}
	data["TANGGAL"] = time.Now().Format("02 January 2006")

	// 5. Ambil jenis surat
	jenisStr, _ := db.GetEditField(dbConn, jid)
	jenis := surat.JenisSurat(jenisStr)

	// 6. Buat PDF (secara asinkron)
	_, err = surat.GenerateAsync(jenis, data, "temp", jid, waClient)
	if err != nil {
		log.Printf("[SURAT-ERROR] %v", err)
		return []string{"Terjadi kesalahan saat memproses surat."}
	}
	// 7. Bersihkan sesi
	db.DeleteDataEntrySession(dbConn, jid)
	db.ClearTemporaryByPrefix(jid + "_")
	db.ClearTemporaryByPrefix("surat_fields_" + jid)
	db.ClearTemporaryByPrefix("surat_field_now_" + jid)
	db.ClearTemporaryByPrefix("surat_valid_nik_" + jid)

	return []string{
		fmt.Sprintf("Surat *%s* Anda sedang diproses dan akan segera dikirimkan. Harap tunggu...", surat.NamaSuratmap[jenis]),
		// fmt.Sprintf("(Debug: File LaTeX dibuat di %s)", path),
	}
}

func FastTrackDataEntry(dbConn *sqlx.DB, jid string) error {
	// Add default IDs for all lookup tables
	query := `
        UPDATE data_entry_sessions 
        SET alamat = 'Jl. Contoh No. 1',
            dusun = 'Dusun Contoh',
            rw = '001',
            rt = '002',
            nama = 'Budi Santoso',
            no_kk = '1234567890123456',
            nik = '9876543210987654',
            sex_id = 1,
            tempat_lahir = 'Jakarta',
            tanggal_lahir = '1990-01-01'::date,
            agama_id = 1,
            pendidikan_kk_id = 1,
            pendidikan_sedang_id = 1,
            pekerjaan_id = 1,
            status_kawin_id = 1,
            kk_level_id = 1,
            warganegara_id = 1,
            nik_ayah = '1234567890123456',
            nama_ayah = 'John',
            nik_ibu = '1234567890123456',
            nama_ibu = 'Jane',
            golongan_darah_id = 1,
            akta_lahir = 'AKTA123',
            dokumen_passport = 'PASPOR123',
            tanggal_akhir_passport = '2025-12-31'::date,
            dokumen_kitas = 'KITAS123',
            akta_perkawinan = 'AKTA456',
            tanggal_perkawinan = '1990-01-01'::date,
            akta_perceraian = 'AKTA789',
            tanggal_perceraian = '1990-01-01'::date,
            cacat_id = 1,
            cara_kb_id = 1,
            hamil_id = 1,
            ktp_el_id = 2,             -- Set to KTP-EL
            status_rekam_id = 1,
            alamat_sekarang = 'Alamat Sekarang',
            status_dasar_id = 1,
            suku_id = 1,
            tag_card = 'TAG123',
            id_asuransi_id = 1,
            no_asuransi = '23',
            current_step = 41
        WHERE jid = $1`

	_, err := dbConn.Exec(query, jid)
	return err
}

func validateInput(text string, step Step) (interface{}, error) {
	text = strings.TrimSpace(text)

	// Special validation for empty input
	if text == "" {
		return nil, fmt.Errorf("input tidak boleh kosong\n\n%s", step.Question)
	}

	// Handle options validation first
	if step.Options != nil {
		choice, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("⚠️ mohon pilih nomor yang tersedia dari daftar di bawah\n\n%s",
				formatQuestionWithOptions(step.Question, step.Options))
		}

		// Validate against available options
		if _, valid := step.Options[choice]; !valid {
			return nil, fmt.Errorf("⚠️ pilihan tidak valid. mohon pilih nomor yang tersedia dari daftar di bawah\n\n%s",
				formatQuestionWithOptions(step.Question, step.Options))
		}

		return choice, nil
	}

	// Field-specific validations
	switch step.Field {
	case "nik", "no_kk", "nik_ayah", "nik_ibu":
		if len(text) != 16 {
			return nil, fmt.Errorf("⚠️ %s harus 16 digit\npanjang input saat ini: %d digit\n\n%s",
				strings.ToLower(step.Field), len(text), formatQuestion(step))
		}
		if _, err := strconv.ParseInt(text, 10, 64); err != nil {
			return nil, fmt.Errorf("⚠️ %s hanya boleh berisi angka\n\n%s",
				strings.ToLower(step.Field), formatQuestion(step))
		}
		return text, nil
	case "rt", "rw":
		if len(text) != 3 {
			return nil, fmt.Errorf("⚠️ %s harus 3 digit\ncontoh format yang benar: 001\n\n%s",
				strings.ToUpper(step.Field), step.Question)
		}
		if _, err := strconv.Atoi(text); err != nil {
			return nil, fmt.Errorf("⚠️ %s hanya boleh berisi angka\n\n%s",
				strings.ToUpper(step.Field), step.Question)
		}
	case "nama", "nama_ayah", "nama_ibu":
		if len(text) < 2 {
			return nil, fmt.Errorf("⚠️ %s terlalu pendek\nminimal 2 karakter\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
		if strings.ContainsAny(text, "0123456789") {
			return nil, fmt.Errorf("⚠️ %s tidak boleh mengandung angka\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
		// Check for special characters
		if strings.ContainsAny(text, "!@#$%^&*()_+=[]{};:'\"\\|,.<>/?") {
			return nil, fmt.Errorf("⚠️ %s tidak boleh mengandung karakter khusus\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
	case "alamat", "alamat_sekarang":
		if len(text) < 5 {
			return nil, fmt.Errorf("⚠️ %s terlalu pendek\nminimal 5 karakter\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
	case "tempat_lahir":
		if len(text) < 3 {
			return nil, fmt.Errorf("⚠️ tempat lahir terlalu pendek\nminimal 3 karakter\n\n%s",
				step.Question)
		}
		if strings.ContainsAny(text, "0123456789") {
			return nil, fmt.Errorf("⚠️ tempat lahir tidak boleh mengandung angka\n\n%s",
				step.Question)
		}
	case "akta_lahir", "dokumen_passport", "dokumen_kitas", "akta_perkawinan", "akta_perceraian":
		if len(text) < 3 {
			return nil, fmt.Errorf("⚠️ nomor dokumen terlalu pendek\nminimal 3 karakter\n\n%s",
				step.Question)
		}
		// Allow alphanumeric but no special chars except dash and forward slash
		if !regexp.MustCompile(`^[a-zA-Z0-9-/]+$`).MatchString(text) {
			return nil, fmt.Errorf("⚠️ nomor dokumen hanya boleh berisi huruf, angka, dash (-) dan garis miring (/)\n\n%s",
				step.Question)
		}
	}

	return text, nil
}

func formatQuestion(step Step) string {
	if step.Options != nil {
		return formatQuestionWithOptions(step.Question, step.Options)
	}

	// Add hints for specific fields
	switch step.Field {
	case "nik", "no_kk":
		return fmt.Sprintf("%s\n(Masukkan 16 digit angka)", step.Question)
	case "rt", "rw":
		return fmt.Sprintf("%s\n(Masukkan 3 digit angka, contoh: 001)", step.Question)
	case "tanggal_lahir", "tanggal_akhir_passport", "tanggal_perkawinan", "tanggal_perceraian":
		return fmt.Sprintf("%s\n(Format: DD-MM-YYYY, contoh: 01-12-2024)", step.Question)
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
