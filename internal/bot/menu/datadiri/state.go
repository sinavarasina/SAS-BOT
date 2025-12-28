package datadiri

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Step struct {
	Question string
	Field    string
	IsDate   bool
	Options  map[int]string
}

var Steps = map[int]Step{
	// Data Inti (1-19)
	1:  {"Masukkan Dusun:", "dusun", false, nil},
	2:  {"Masukkan RT (contoh: 001):", "rt", false, nil},
	3:  {"Masukkan Nama:", "nama", false, nil},
	4:  {"Masukkan No. KK (16 digit):", "no_kk", false, nil},
	5:  {"Masukkan NIK (16 digit):", "nik", false, nil},
	6:  {"Pilih Jenis Kelamin:", "sex_id", false, loadOptions("6_sex.json", "sex")},
	7:  {"Masukkan Tempat Lahir:", "tempat_lahir", false, nil},
	8:  {"Masukkan Tanggal Lahir (DD-MM-YYYY):", "tanggal_lahir", true, nil},
	9:  {"Pilih Agama:", "agama_id", false, loadOptions("9_agama.json", "agama")},
	10: {"Pilih Pendidikan Dalam KK:", "pendidikan_kk_id", false, loadOptions("10_pendidikan_kk.json", "pendidikan_kk")},
	11: {"Pilih Pendidikan Sedang Ditempuh:", "pendidikan_sedang_id", false, loadOptions("11_pendidikan_sedang.json", "pendidikan_sedang")},
	12: {"Pilih Pekerjaan:", "pekerjaan_id", false, loadOptions("12_pekerjaan.json", "pekerjaan")},
	13: {"Pilih Status Kawin:", "status_kawin_id", false, loadOptions("13_status_kawin.json", "status_kawin")},
	14: {"Pilih Status Hubungan Dalam KK:", "kk_level_id", false, loadOptions("14_kk_level.json", "kk_level")},
	15: {"Pilih Warganegara:", "warganegara_id", false, loadOptions("15_warganegara.json", "warganegara")},
	16: {"Masukkan Nama Ayah:", "nama_ayah", false, nil},
	17: {"Masukkan Nama Ibu:", "nama_ibu", false, nil},
	18: {"Pilih Status Dasar:", "status_dasar_id", false, loadOptions("18_status_dasar.json", "status_dasar")},
	19: {"Pilih Suku:", "suku_id", false, loadOptions("19_suku.json", "suku")},

	// Data Lanjutan (20-39)
	20: {"Masukkan NIK Ayah (16 digit, '0' jika tidak ada):", "nik_ayah", false, nil},
	21: {"Masukkan NIK Ibu (16 digit, '0' jika tidak ada):", "nik_ibu", false, nil},
	22: {"Pilih Golongan Darah:", "golongan_darah_id", false, loadOptions("22_golongan_darah.json", "golongan_darah")},
	23: {"Masukkan No. Akta Kelahiran ('-' jika tidak ada):", "akta_lahir", false, nil},
	24: {"Masukkan No. Dokumen Passport ('-' jika tidak ada):", "dokumen_passport", false, nil},
	25: {"Masukkan Tanggal Akhir Passport (DD-MM-YYYY, '-' jika tidak ada):", "tanggal_akhir_passport", true, nil},
	26: {"Masukkan No. Dokumen KITAS/KITAP ('-' jika tidak ada):", "dokumen_kitas", false, nil},
	27: {"Masukkan No. Akta Perkawinan ('-' jika tidak ada):", "akta_perkawinan", false, nil},
	28: {"Masukkan Tanggal Perkawinan (DD-MM-YYYY, '-' jika tidak ada):", "tanggal_perkawinan", true, nil},
	29: {"Masukkan No. Akta Perceraian ('-' jika tidak ada):", "akta_perceraian", false, nil},
	30: {"Masukkan Tanggal Perceraian (DD-MM-YYYY, '-' jika tidak ada):", "tanggal_perceraian", true, nil},
	31: {"Pilih Jenis Cacat:", "cacat_id", false, loadOptions("31_cacat.json", "cacat")},
	32: {"Pilih Cara KB:", "cara_kb_id", false, loadOptions("32_cara_kb.json", "cara_kb")},
	33: {"Pilih Status Kehamilan (bagi wanita):", "hamil_id", false, loadOptions("33_hamil.json", "hamil")},
	34: {"Pilih Status KTP-el:", "ktp_el_id", false, loadOptions("34_ktp_el.json", "ktp_el")},
	35: {"Pilih Status Rekam:", "status_rekam_id", false, loadOptions("35_status_rekam.json", "status_rekam")},
	36: {"Masukkan Alamat Sekarang (jika beda):", "alamat_sekarang", false, nil},
	37: {"Masukkan No. Tag Card ('-' jika tidak ada):", "tag_card", false, nil},
	38: {"Pilih Jenis Asuransi:", "id_asuransi_id", false, loadOptions("38_asuransi.json", "id_asuransi")},
	39: {"Masukkan No. Asuransi ('-' jika tidak ada):", "no_asuransi", false, nil},
}

// Konstanta Langkah (agar mudah dirujuk)
const (
	STEP_DUSUN = 1
	STEP_RT    = 2
	STEP_NAMA  = 3
	STEP_NO_KK = 4
	STEP_NIK   = 5
	STEP_SEX   = 6
	// ... (bisa ditambahkan jika perlu)
	STEP_SUKU        = 19
	STEP_NIK_AYAH    = 20
	STEP_NO_ASURANSI = 39
)

func loadOptions(fileName, key string) map[int]string {
	options := make(map[int]string)
	filePath := filepath.Join("json", fileName)
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Gagal membaca %s: %v\n", filePath, err)
		return options
	}

	var data map[string][]map[string]any
	if err := json.Unmarshal(file, &data); err == nil {
		if records, ok := data[key]; ok {
			for _, opt := range records {
				id := int(opt["id"].(float64))
				nama := opt["nama"].(string)
				options[id] = nama
			}
		}
	} else {
		var records []map[string]any
		if errAlt := json.Unmarshal(file, &records); errAlt == nil {
			keyGuess := strings.Split(fileName, "_")[1]
			keyGuess = strings.TrimSuffix(keyGuess, ".json")
			if key == keyGuess || (key == "dusun" && fileName == "1_dusun.json") || (key == "sex" && fileName == "6_sex.json") || (key == "agama" && fileName == "9_agama.json") || (key == "pendidikan_kk" && fileName == "10_pendidikan_kk.json") || (key == "pendidikan_sedang" && fileName == "11_pendidikan_sedang.json") || (key == "pekerjaan" && fileName == "12_pekerjaan.json") || (key == "status_kawin" && fileName == "13_status_kawin.json") || (key == "kk_level" && fileName == "14_kk_level.json") || (key == "warganegara" && fileName == "15_warganegara.json") || (key == "status_dasar" && fileName == "18_status_dasar.json") || (key == "suku" && fileName == "19_suku.json") {
				for _, opt := range records {
					id := int(opt["id"].(float64))
					nama := opt["nama"].(string)
					options[id] = nama
				}
			}
		} else {
			log.Printf("[ERROR] Gagal parse JSON %s: %v", fileName, err)
		}
	}
	return options
}
func FormatQuestion(step Step) string {
	if step.Options != nil {
		var builder strings.Builder
		builder.WriteString(step.Question)

		ids := make([]int, 0, len(step.Options))
		for id := range step.Options {
			ids = append(ids, id)
		}
		sort.Ints(ids)

		for _, id := range ids {
			builder.WriteString(fmt.Sprintf("\n%d. %s", id, step.Options[id]))
		}
		return builder.String()
	}

	switch step.Field {
	case "nik", "no_kk":
		return fmt.Sprintf("%s\n(Masukkan 16 digit angka)", step.Question)
	case "rt":
		return fmt.Sprintf("%s\n(Masukkan 3 digit angka, contoh: 001)", step.Question)
	case "tanggal_lahir":
		return fmt.Sprintf("%s\n(Format: DD-MM-YYYY, contoh: 01-12-2024)", step.Question)
	}
	return step.Question
}
