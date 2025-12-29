package surat

import (
	"fmt"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"slices"
	"strings"
	"time"
)

type JenisSurat string

const (
	DOMISILI       JenisSurat = "sk_domisili.tex"
	USAHA          JenisSurat = "sk_usaha.tex"
	SKTM_UMUM      JenisSurat = "sktm_umum.tex"
	KEMATIAN       JenisSurat = "sk_kematian.tex"
	IJIN_KELUARGA  JenisSurat = "sk_ijin_keluarga.tex"
	IZIN_KERAMAIAN JenisSurat = "sk_izin_keramaian.tex"
	KELAHIRAN      JenisSurat = "sk_kelahiran.tex"
	SKCK           JenisSurat = "sk_skck.tex"
	BEDA_NAMA      JenisSurat = "sk_beda_nama.tex"
)

var JenisSuratMap = map[string]JenisSurat{
	"1": DOMISILI,
	"2": USAHA,
	"3": SKTM_UMUM,
	"4": KEMATIAN,
	"5": IJIN_KELUARGA,
	"6": IZIN_KERAMAIAN,
	"7": KELAHIRAN,
	"8": SKCK,
	"9": BEDA_NAMA,
}

var NamaSuratmap = map[JenisSurat]string{
	DOMISILI:       "Surat Keterangan Domisili",
	USAHA:          "Surat Keterangan Usaha",
	SKTM_UMUM:      "SKTM Umum",
	KEMATIAN:       "Surat Keterangan Kematian",
	IJIN_KELUARGA:  "Surat Keterangan Ijin Keluarga",
	IZIN_KERAMAIAN: "Surat Keterangan Izin Keramaian",
	KELAHIRAN:      "Surat Keterangan Kelahiran",
	SKCK:						"Surat Pengantar SKCK",
	BEDA_NAMA:			"Surat Keterangan Beda Nama",
}

// BaseFields adalah data yang kita AMBIL OTOMATIS dari DB
var BaseFields = []string{
	"NAMA", "TTL", "JK", "AGAMA", "NIK", "PEKERJAAN", "ALAMAT",
}

// SuratFields HANYA berisi field TAMBAHAN yang perlu ditanyakan.
var SuratFields = map[JenisSurat][]string{
	DOMISILI:  {"ALASANPERLU"},
	USAHA:     {"TTLnU", "ALAMATDOM", "DUSUN"},
	SKTM_UMUM: {"TTLnU", "ALASANPERLU"},
	KEMATIAN: {
		"FNAMAnAL", "BINoBINTI", "TTLnU", "AGAMA", "PEKERJAAN", "POSISITERAKHIR",
		"HARI", "TGL", "JAM", "TEMPAT", "ALASANMENINGGAL",
	},
	IJIN_KELUARGA: {
		"HUBUNGAN_WALI", "NAMA_CPMI", "TTL_CPMI", "NIK_CPMI", "JK_CPMI", 
		"PEKERJAAN_CPMI", "ALAMAT_CPMI", "NEGARA_TUJUAN",
	},
	IZIN_KERAMAIAN: {
		"HP_PEMOHON", "JENIS_KESENIAN", "NAMA_PIMPINAN_HIBURAN", "HP_PIMPINAN", 
		"ACARA_HAJATAN", "HARI_ACARA", "TGL_ACARA", "JAM_ACARA", "LOKASI_ACARA",
		"NAMA_BABIN", // DITAMBAHKAN: Agar sesuai template source 11
	},
	KELAHIRAN: {
		"HARI_LAHIR", "TGL_LAHIR", "TEMPAT_LAHIR", "JK_ANAK", "NAMA_ANAK", "ANAK_KE",
		"NAMA_IBU", "TTL_IBU", "AGAMA_IBU", "ALAMAT_IBU",
		"NAMA_AYAH", "TTL_AYAH", "AGAMA_AYAH", "ALAMAT_AYAH",
	},
	SKCK: {
		"SUKU_BANGSA", "NAMA_ORTU", "TANDA_ISTIMEWA", "KEPERLUAN", "CATATAN_SKCK",
	},
	BEDA_NAMA: {
		"SUMBER_DATA_1", // Sumber data BENAR (misal: KTP)
		"SUMBER_DATA_2", // Sumber data SALAH (misal: Ijazah)
		// Kita asumsikan Data 1 diambil dari DB user (Benar), kita tanya Data 2 (Salah/Beda)
		"NAMA_2", "TTL_2", "NIK_2", "PEKERJAAN_2", "ALAMAT_2",
	},
}

// FieldPrompts adalah daftar pertanyaan untuk setiap field
var FieldPrompts = map[string]string{
	// Existing
	"ALASANPERLU":    "Tuliskan alasan Anda membutuhkan surat ini:",
	"ALAMATDOM":      "Tuliskan alamat domisili atau lokasi usaha:",
	"DUSUN":          "Tuliskan nama dusun tempat tinggal Anda:",
	"TTLnU":          "Tuliskan tempat & tanggal lahir atau umur (misal: Bandar Lampung, 19 Juni 1975 / 49 tahun):",

	//kematian 
	"FNAMAnAL":       "Tuliskan nama lengkap almarhum/almarhumah:",
	"BINoBINTI":      "Tuliskan Bin/Binti (nama ayah/almarhum):",
	"POSISITERAKHIR": "Tuliskan tempat tinggal atau posisi terakhir almarhum:",
	"HARI":           "Tuliskan hari meninggalnya:",
	"TGL":            "Tuliskan tanggal meninggalnya (misal: 04 November 2025):",
	"JAM":            "Tuliskan jam meninggalnya:",
	"TEMPAT":         "Tuliskan tempat meninggalnya:",
	"ALASANMENINGGAL": "Tuliskan penyebab meninggalnya:",
	"AGAMA":          "Agama almarhum:",
	"PEKERJAAN":      "Pekerjaan almarhum:",

	// Prompts untuk IJIN_KELUARGA
	"HUBUNGAN_WALI":  "Hubungan Anda dengan CPMI (misal: CPMI itu siapa anda? Suami/Istri/Orang Tua):",
	"NAMA_CPMI":      "Nama lengkap CPMI yang akan berangkat:",
	"TTL_CPMI":       "Tempat & Tanggal Lahir CPMI:",
	"NIK_CPMI":       "NIK CPMI:",
	"JK_CPMI":        "Jenis Kelamin CPMI (Laki-laki/Perempuan):",
	"PEKERJAAN_CPMI": "Pekerjaan CPMI saat ini:",
	"ALAMAT_CPMI":    "Alamat lengkap CPMI:",
	"NEGARA_TUJUAN":  "Negara tujuan bekerja:",

	// Prompts untuk IZIN_KERAMAIAN
	"HP_PEMOHON":            "Nomor HP Pemohon (Anda):",
	"JENIS_KESENIAN":        "Jenis hiburan/kesenian yang akan digelar:",
	"NAMA_PIMPINAN_HIBURAN": "Nama pimpinan hiburan/orkes:",
	"HP_PIMPINAN":           "Nomor HP pimpinan hiburan:",
	"ACARA_HAJATAN":         "Dalam rangka acara apa (misal: Khitanan/Pernikahan):",
	"HARI_ACARA":            "Hari pelaksanaan acara:",
	"TGL_ACARA":             "Tanggal pelaksanaan acara:",
	"JAM_ACARA":             "Jam pelaksanaan acara (WIB):",
	"LOKASI_ACARA":          "Lokasi lengkap acara:",
	"NAMA_BABIN":            "Nama Babinkamtibmas yang bertugas:", // Prompt baru

	// Prompts untuk KELAHIRAN
	"HARI_LAHIR":   "Hari kelahiran anak:",
	"TGL_LAHIR":    "Tanggal kelahiran anak:",
	"TEMPAT_LAHIR": "Tempat kelahiran anak:",
	"JK_ANAK":      "Jenis kelamin anak:",
	"NAMA_ANAK":    "Nama lengkap anak:",
	"ANAK_KE":      "Anak ke berapa:",
	"NAMA_IBU":     "Nama lengkap Ibu:",
	"TTL_IBU":      "Tempat & Tanggal Lahir Ibu:",
	"AGAMA_IBU":    "Agama Ibu:",
	"ALAMAT_IBU":   "Alamat Ibu:",
	"NAMA_AYAH":    "Nama lengkap Ayah:",
	"TTL_AYAH":     "Tempat & Tanggal Lahir Ayah:",
	"AGAMA_AYAH":   "Agama Ayah:",
	"ALAMAT_AYAH":  "Alamat Ayah:",

	// --- SKCK ---
	"SUKU_BANGSA":    "Suku Bangsa (misal: Indonesia/Jawa):",
	"NAMA_ORTU":      "Nama Orang Tua:",
	"TANDA_ISTIMEWA": "Tanda Istimewa / Ciri Fisik (misal: Tahi lalat di pipi / - ):",
	"KEPERLUAN":      "Keperluan pembuatan SKCK (misal: Melamar Pekerjaan):",
	"CATATAN_SKCK":   "Catatan Prilaku (misal: Berkelakuan Baik / Tidak pernah tersangkut pidana):",

	// --- BEDA NAMA ---
	"SUMBER_DATA_1": "Sebutkan Dokumen yang datanya BENAR/SESUAI KTP (misal: KTP & KK):",
	"SUMBER_DATA_2": "Sebutkan Dokumen yang datanya SALAH/BERBEDA (misal: Ijazah / Akta Cerai):",
	"NAMA_2":        "Tuliskan NAMA yang tertulis pada dokumen SALAH tersebut:",
	"TTL_2":         "Tuliskan TEMPAT TANGGAL LAHIR pada dokumen SALAH tersebut:",
	"NIK_2":         "Tuliskan NIK pada dokumen SALAH tersebut (jika ada, atau strip '-'):",
	"PEKERJAAN_2":   "Tuliskan PEKERJAAN pada dokumen SALAH tersebut:",
	"ALAMAT_2":      "Tuliskan ALAMAT pada dokumen SALAH tersebut:",
}

// GetPrompt mengambil pertanyaan untuk field
func GetPrompt(field string) string {
	if prompt, ok := FieldPrompts[field]; ok {
		return prompt
	}
	return fmt.Sprintf("Masukkan data untuk *%s*:", field)
}

// BuildDataMap membuat map[string]string (data auto-fill)
func BuildDataMap(data db.DataPenduduk) map[string]string {
	ttl := fmt.Sprintf("%s, %s", data.TempatLahir.String, db.FormatDate(data.TanggalLahir))
	jk := "Laki-laki"
	if data.SexID.Int64 == 2 {
		jk = "Perempuan"
	}
	// Alamat Disesuaikan (Tanpa Alamat & RW)
	alamatLengkap := fmt.Sprintf("Dusun %s, RT %s", data.Dusun.String, data.RT.String)

	return map[string]string{
		// Base Keys
		"NAMA":      data.Nama.String,
		"TTL":       ttl,
		"TTLnU":     ttl,
		"JK":        jk,
		"AGAMA":     data.AgamaNama.String,
		"NIK":       data.NIK.String,
		"PEKERJAAN": data.PekerjaanNama.String,
		"ALAMAT":    alamatLengkap,
		"TANGGAL":   time.Now().Format("02 January 2006"),
		
		// Auto-fill Mapping untuk Template Izin Keramaian (Pemohon = User)
		"NAMA_PEMOHON":      data.Nama.String,
		"TTL_PEMOHON":       ttl,
		"PEKERJAAN_PEMOHON": data.PekerjaanNama.String,
		"NIK_PEMOHON":       data.NIK.String,
		"ALAMAT_PEMOHON":    alamatLengkap,

		// Auto-fill Mapping untuk Template Ijin Keluarga (Wali = User)
		"NAMA_WALI":      data.Nama.String,
		"JK_WALI":        jk,
		"TTL_WALI":       ttl,
		"NIK_WALI":       data.NIK.String,
		"PEKERJAAN_WALI": data.PekerjaanNama.String,
		"ALAMAT_WALI":    alamatLengkap,

		//Autofill untuk temlate beda nama
		"NAMA_1":      data.Nama.String,
		"TTL_1":       ttl,
		"NIK_1":       data.NIK.String,
		"PEKERJAAN_1": data.PekerjaanNama.String,
		"ALAMAT_1":    alamatLengkap,
	}
}

func GetFieldList(data db.DataPenduduk, jenis JenisSurat) []string {
	// Jika jenis surat tidak perlu pengecekan BaseFields (semua manual atau list fixed), return langsung
	if jenis == KEMATIAN || jenis == IJIN_KELUARGA || jenis == IZIN_KERAMAIAN || jenis == KELAHIRAN || jenis == SKCK || jenis == BEDA_NAMA {
		return SuratFields[jenis]
	}

	neededFields := []string{}
	allExtraFields := SuratFields[jenis]

	for _, field := range allExtraFields {
		if !slices.Contains(BaseFields, field) {
			neededFields = append(neededFields, field)
		}
	}

	return neededFields
}

// NextField mencari field selanjutnya dalam daftar
func NextField(fields []string, current string) string {
	for i, f := range fields {
		if strings.TrimSpace(f) == strings.TrimSpace(current) && i+1 < len(fields) {
			return strings.TrimSpace(fields[i+1])
		}
	}
	return ""
}
