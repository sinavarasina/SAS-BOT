package datadiri

import (
	"encoding/json"
	"fmt"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Step mendefinisikan satu langkah dalam alur
type Step struct {
	Question string
	Field    string
	IsDate   bool
	Options  map[int]string
}

// Steps FSM (Finite State Machine) untuk 19 langkah
var Steps = map[int]Step{
	common.STEP_DUSUN:        {"Masukkan Dusun:", "dusun", false, nil},
	common.STEP_RT:           {"Masukkan RT (contoh: 001):", "rt", false, nil},
	common.STEP_NAMA:         {"Masukkan Nama:", "nama", false, nil},
	common.STEP_NO_KK:        {"Masukkan No. KK (16 digit):", "no_kk", false, nil},
	common.STEP_NIK:          {"Masukkan NIK (16 digit):", "nik", false, nil},
	common.STEP_SEX:          {"Pilih Jenis Kelamin:", "sex_id", false, loadOptions("6_sex.json", "sex")},
	common.STEP_TEMPAT_LAHIR: {"Masukkan Tempat Lahir:", "tempat_lahir", false, nil},
	common.STEP_TANGGAL_LAHIR:  {"Masukkan Tanggal Lahir (DD-MM-YYYY):", "tanggal_lahir", true, nil},
	common.STEP_AGAMA:        {"Pilih Agama:", "agama_id", false, loadOptions("9_agama.json", "agama")},
	common.STEP_PENDIDIKAN_KK: {"Pilih Pendidikan Dalam KK:", "pendidikan_kk_id", false, loadOptions("10_pendidikan_kk.json", "pendidikan_kk")},
	common.STEP_PENDIDIKAN_SEDANG: {"Pilih Pendidikan Sedang Ditempuh:", "pendidikan_sedang_id", false, loadOptions("11_pendidikan_sedang.json", "pendidikan_sedang")},
	common.STEP_PEKERJAAN:      {"Pilih Pekerjaan:", "pekerjaan_id", false, loadOptions("12_pekerjaan.json", "pekerjaan")},
	common.STEP_STATUS_KAWIN: {"Pilih Status Kawin:", "status_kawin_id", false, loadOptions("13_status_kawin.json", "status_kawin")},
	common.STEP_KK_LEVEL:     {"Pilih Status Hubungan Dalam KK:", "kk_level_id", false, loadOptions("14_kk_level.json", "kk_level")},
	common.STEP_WARGANEGARA:  {"Pilih Warganegara:", "warganegara_id", false, loadOptions("15_warganegara.json", "warganegara")},
	common.STEP_NAMA_AYAH:    {"Masukkan Nama Ayah:", "nama_ayah", false, nil},
	common.STEP_NAMA_IBU:     {"Masukkan Nama Ibu:", "nama_ibu", false, nil},
	common.STEP_STATUS_DASAR: {"Pilih Status Dasar:", "status_dasar_id", false, loadOptions("18_status_dasar.json", "status_dasar")},
	common.STEP_SUKU:         {"Pilih Suku:", "suku_id", false, loadOptions("19_suku.json", "suku")},
}

// loadOptions memuat data dari file JSON (sama seperti 'loadJSONOptions' lama Anda)
func loadOptions(fileName, key string) map[int]string {
	options := make(map[int]string)
	filePath := filepath.Join("json", fileName)
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Gagal membaca %s: %v\n", filePath, err)
		return options
	}

	var data map[string][]map[string]interface{}
	if err := json.Unmarshal(file, &data); err == nil {
		if records, ok := data[key]; ok {
			for _, opt := range records {
				id := int(opt["id"].(float64))
				nama := opt["nama"].(string)
				options[id] = nama
			}
		}
	}
	return options
}

// FormatQuestion memformat pertanyaan untuk pengguna
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
