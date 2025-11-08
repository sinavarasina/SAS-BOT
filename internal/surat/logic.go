package surat

import (
	"fmt"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db"
)

// JenisSuratMap memetakan input teks (pilihan user) ke tipe surat
var JenisSuratMap = map[string]JenisSurat{
	"1": DOMISILI,
	"2": USAHA,
	"3": SKTM_UMUM,
	"4": SKTM_TANGGUNGAN,
	"5": KEMATIAN,
}

// NamaSuratmap memetakan tipe surat ke nama yang ramah
var NamaSuratmap = map[JenisSurat]string{
	DOMISILI:        "Surat Keterangan Domisili",
	USAHA:           "Surat Keterangan Usaha",
	SKTM_UMUM:       "SKTM Umum",
	SKTM_TANGGUNGAN: "SKTM Tanggungan",
	KEMATIAN:        "Surat Keterangan Kematian",
}

// FieldPrompts adalah daftar pertanyaan untuk setiap field
var FieldPrompts = map[string]string{
	"NAMA":      "Tuliskan nama lengkap Anda:",
	"TTL":       "Tuliskan tempat dan tanggal lahir Anda (misal: Bandung, 1 Januari 1990):",
	"JK":        "Apa jenis kelamin Anda? (L/P):",
	"AGAMA":     "Apa agama Anda?",
	"NIK":       "Tuliskan NIK Anda:",
	"PEKERJAAN": "Apa pekerjaan Anda?",
	"ALAMAT":    "Tuliskan alamat lengkap Anda:",

	"ALASANPERLU": "Tuliskan alasan Anda membutuhkan surat ini:",
	"ALAMATDOM":   "Tuliskan alamat domisili yang dimaksud:",
	"DUSUN":       "Tuliskan dusun tempat tinggal Anda:",

	"FNAMAnAL":        "Tuliskan nama lengkap almarhum/almarhumah:",
	"BINoBINTI":       "Tuliskan Bin/Binti (nama ayah/almarhum):",
	"TTLnU":           "Tuliskan tempat & tanggal lahir almarhum/almarhumah:",
	"POSISITERAKHIR":  "Tuliskan posisi terakhir almarhum (misal: PNS, Petani, dst):",
	"HARI":            "Tuliskan hari meninggalnya:",
	"TGL":             "Tuliskan tanggal meninggalnya (misal: 04 November 2025):",
	"JAM":             "Tuliskan jam meninggalnya:",
	"TEMPAT":          "Tuliskan tempat meninggalnya:",
	"ALASANMENINGGAL": "Tuliskan penyebab meninggalnya:",

	"NAMA.P":      "Nama orang tua atau penanggung:",
	"JK.P":        "Jenis kelamin orang tua/penanggung (L/P):",
	"TTL.P":       "Tempat & tanggal lahir orang tua/penanggung:",
	"NIK.P":       "NIK orang tua/penanggung:",
	"PEKERJAAN.P": "Pekerjaan orang tua/penanggung:",
	"AGAMA.P":     "Agama orang tua/penanggung:",
	"STATUS.P":    "Status hubungan dengan anak (misal: Ayah/Ibu):",
	"ALAMAT.P":    "Alamat orang tua/penanggung:",
	"NAMA.C":      "Nama anak:",
	"JK.C":        "Jenis kelamin anak (L/P):",
	"TTL.C":       "Tempat & tanggal lahir anak:",
	"NIK.C":       "NIK anak:",
	"PEKERJAAN.C": "Pekerjaan anak:",
	"AGAMA.C":     "Agama anak:",
	"ALAMAT.C":    "Alamat anak:",
}

// GetFieldList menentukan field apa saja yang *perlu ditanyakan* ke user
// (Menggunakan BaseFields dan SuratFields dari model.go)
func GetFieldList(data db.DataPenduduk, jenis JenisSurat) []string {
	if jenis == KEMATIAN || jenis == SKTM_TANGGUNGAN {
		return SuratFields[jenis]
	}

	neededFields := []string{}
	allExtraFields := SuratFields[jenis]

	for _, field := range allExtraFields {
		isBase := false
		for _, base := range BaseFields {
			if field == base {
				isBase = true
				break
			}
		}
		if !isBase {
			neededFields = append(neededFields, field)
		}
	}
	return neededFields
}

// BuildDataMap membuat map[string]string untuk template LaTeX
func BuildDataMap(data db.DataPenduduk) map[string]string {
	ttl := fmt.Sprintf("%s, %s", data.TempatLahir.String, db.FormatDate(data.TanggalLahir))
	jk := "Laki-laki"
	if data.SexID.Int64 == 2 {
		jk = "Perempuan"
	}

	return map[string]string{
		"NAMA":      data.Nama.String,
		"TTL":       ttl,
		"JK":        jk,
		// "AGAMA":     data.Agama.String, // Asumsi Anda menambah AgamaNama di struct DataPenduduk
		"NIK":       data.NIK.String,
		// "PEKERJAAN": data.Pekerjaan.String, // Asumsi Anda menambah PekerjaanNama
		"ALAMAT":    fmt.Sprintf("%s, Dusun %s, RT/RW %s/%s", data.Alamat.String, data.Dusun.String, data.RT.String, data.RW.String),
		"TANGGAL":   time.Now().Format("02 January 2006"),
	}
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

// GetPrompt mengambil pertanyaan untuk field
func GetPrompt(field string) string {
	if prompt, ok := FieldPrompts[field]; ok {
		return prompt
	}
	return fmt.Sprintf("Masukkan data untuk *%s*:", field)
}
