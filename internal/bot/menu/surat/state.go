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
	DOMISILI        JenisSurat = "sk_domisili.tex"
	USAHA           JenisSurat = "sk_usaha.tex"
	SKTM_UMUM       JenisSurat = "sktm_umum.tex"
	SKTM_TANGGUNGAN JenisSurat = "sktm_tanggungan.tex"
	KEMATIAN        JenisSurat = "sk_kematian.tex"
)

var JenisSuratMap = map[string]JenisSurat{
	"1": DOMISILI,
	"2": USAHA,
	"3": SKTM_UMUM,
	"4": SKTM_TANGGUNGAN,
	"5": KEMATIAN,
}

var NamaSuratmap = map[JenisSurat]string{
	DOMISILI:        "Surat Keterangan Domisili",
	USAHA:           "Surat Keterangan Usaha",
	SKTM_UMUM:       "SKTM Umum",
	SKTM_TANGGUNGAN: "SKTM Tanggungan",
	KEMATIAN:        "Surat Keterangan Kematian",
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
	SKTM_TANGGUNGAN: {
		"NAMA.P", "JK.P", "TTL.P", "NIK.P", "PEKERJAAN.P", "AGAMA.P", "STATUS.P", "ALAMAT.P",
		"NAMA.C", "JK.C", "TTL.C", "NIK.C", "PEKERJAAN.C", "AGAMA.C", "ALAMAT.C",
		"ALASANPERLU",
	},
	KEMATIAN: {
		"FNAMAnAL", "BINoBINTI", "TTLnU", "AGAMA", "PEKERJAAN", "POSISITERAKHIR",
		"HARI", "TGL", "JAM", "TEMPAT", "ALASANMENINGGAL",
	},
}

// FieldPrompts adalah daftar pertanyaan untuk setiap field
var FieldPrompts = map[string]string{
	"ALASANPERLU": "Tuliskan alasan Anda membutuhkan surat ini:",
	"ALAMATDOM":   "Tuliskan alamat domisili atau lokasi usaha:",
	"DUSUN":       "Tuliskan nama dusun tempat tinggal Anda:",
	"TTLnU":       "Tuliskan tempat & tanggal lahir atau umur (misal: Bandar Lampung, 19 Juni 1975 / 49 tahun):",

	// (Prompt untuk Kematian & SKTM Tanggungan)
	"FNAMAnAL":        "Tuliskan nama lengkap almarhum/almarhumah:",
	"BINoBINTI":       "Tuliskan Bin/Binti (nama ayah/almarhum):",
	"POSISITERAKHIR":  "Tuliskan tempat tinggal atau posisi terakhir almarhum:",
	"HARI":            "Tuliskan hari meninggalnya:",
	"TGL":             "Tuliskan tanggal meninggalnya (misal: 04 November 2025):",
	"JAM":             "Tuliskan jam meninggalnya:",
	"TEMPAT":          "Tuliskan tempat meninggalnya:",
	"ALASANMENINGGAL": "Tuliskan penyebab meninggalnya:",
	"NAMA.P":          "Nama orang tua atau penanggung:",
	"JK.P":            "Jenis kelamin orang tua/penanggung:",
	"TTL.P":           "Tempat & tanggal lahir orang tua/penanggung:",
	"NIK.P":           "NIK orang tua/penanggung:",
	"PEKERJAAN.P":     "Pekerjaan orang tua/penanggung:",
	"AGAMA.P":         "Agama orang tua/penanggung:",
	"STATUS.P":        "Status hubungan dengan anak (misal: Ayah/Ibu):",
	"ALAMAT.P":        "Alamat orang tua/penanggung:",
	"NAMA.C":          "Nama anak:",
	"JK.C":            "Jenis kelamin anak:",
	"TTL.C":           "Tempat & tanggal lahir anak:",
	"NIK.C":           "NIK anak:",
	"PEKERJAAN.C":     "Pekerjaan anak:",
	"AGAMA.C":         "Agama anak:",
	"ALAMAT.C":        "Alamat anak:",
	"AGAMA":           "Agama almarhum:",
	"PEKERJAAN":       "Pekerjaan almarhum:",
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
		"NAMA":      data.Nama.String,
		"TTL":       ttl,
		"TTLnU":     ttl, // (Isi TTLnU dengan data pemohon)
		"JK":        jk,
		"AGAMA":     data.AgamaNama.String,
		"NIK":       data.NIK.String,
		"PEKERJAAN": data.PekerjaanNama.String,
		"ALAMAT":    alamatLengkap,
		"TANGGAL":   time.Now().Format("02 January 2006"),

		"NAMA.P":   data.NamaAyah.String,
		"NIK.P":    data.NikAyah.String,
		"ALAMAT.P": alamatLengkap,
	}
}

func GetFieldList(data db.DataPenduduk, jenis JenisSurat) []string {
	if jenis == KEMATIAN || jenis == SKTM_TANGGUNGAN {
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
