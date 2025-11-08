package surat

import (
	"fmt"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db"
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

var FieldPrompts = map[string]string{
	"NAMA":      "Tuliskan nama lengkap Anda:",
	"TTL":       "Tuliskan tempat dan tanggal lahir Anda (misal: Bandung, 1 Januari 1990):",
	"TTLnU":     "Tuliskan tempat & tanggal lahir atau umur (misal: Bandar Lampung, 10 Juni 1975 / 49 tahun):",
	"JK":        "Apa jenis kelamin Anda? (L/P):",
	"AGAMA":     "Apa agama Anda?",
	"NIK":       "Tuliskan NIK Anda:",
	"PEKERJAAN": "Apa pekerjaan Anda?",
	"ALAMAT":    "Tuliskan alamat lengkap Anda:",

	"ALASANPERLU": "Tuliskan alasan Anda membutuhkan surat ini:",
	"ALAMATDOM":   "Tuliskan alamat domisili atau lokasi usaha:",
	"DUSUN":       "Tuliskan nama dusun tempat tinggal Anda:",

	"FNAMAnAL":        "Tuliskan nama lengkap almarhum/almarhumah:",
	"BINoBINTI":       "Tuliskan Bin/Binti (nama ayah/almarhum):",
	"POSISITERAKHIR":  "Tuliskan tempat tinggal atau posisi terakhir almarhum:",
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

func GetFieldList(_ db.DataPenduduk, jenis JenisSurat) []string {
	if fields, ok := SuratFields[jenis]; ok {
		return fields
	}
	return []string{}
}

func BuildDataMap(data db.DataPenduduk) map[string]string {
	ttl := fmt.Sprintf("%s, %s", data.TempatLahir.String, db.FormatDate(data.TanggalLahir))
	jk := "Laki-laki"
	if data.SexID.Int64 == 2 {
		jk = "Perempuan"
	}

	return map[string]string{
		"NAMA":    data.Nama.String,
		"TTL":     ttl,
		"JK":      jk,
		"NIK":     data.NIK.String,
		"ALAMAT":  fmt.Sprintf("%s, Dusun %s, RT/RW %s/%s", data.Alamat.String, data.Dusun.String, data.RT.String, data.RW.String),
		"TANGGAL": time.Now().Format("02 January 2006"),
	}
}

func NextField(fields []string, current string) string {
	for i, f := range fields {
		if strings.TrimSpace(f) == strings.TrimSpace(current) && i+1 < len(fields) {
			return strings.TrimSpace(fields[i+1])
		}
	}
	return ""
}

func GetPrompt(field string) string {
	if prompt, ok := FieldPrompts[field]; ok {
		return prompt
	}
	return fmt.Sprintf("Masukkan data untuk *%s*:", field)
}
