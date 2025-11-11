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
	"TTLnU":     "Tuliskan tempat & tanggal lahir atau umur (misal: Bandar Lampung, 19 Juni 1975 / 49 tahun):",
	"JK":        "Apa jenis kelamin Anda?:",
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
	"JK.P":        "Jenis kelamin orang tua/penanggung:",
	"TTL.P":       "Tempat & tanggal lahir orang tua/penanggung:",
	"NIK.P":       "NIK orang tua/penanggung:",
	"PEKERJAAN.P": "Pekerjaan orang tua/penanggung:",
	"AGAMA.P":     "Agama orang tua/penanggung:",
	"STATUS.P":    "Status hubungan dengan anak (misal: Ayah/Ibu):",
	"ALAMAT.P":    "Alamat orang tua/penanggung:",
	"NAMA.C":      "Nama anak:",
	"JK.C":        "Jenis kelamin anak:",
	"TTL.C":       "Tempat & tanggal lahir anak:",
	"NIK.C":       "NIK anak:",
	"PEKERJAAN.C": "Pekerjaan anak:",
	"AGAMA.C":     "Agama anak:",
	"ALAMAT.C":    "Alamat anak:",
}

func GetFieldList(data db.DataPenduduk, jenis JenisSurat) []string {
	// Surat Kematian dan SKTM Tanggungan tidak pakai data NIK, tanya semua
	if jenis == KEMATIAN || jenis == SKTM_TANGGUNGAN {
		return SuratFields[jenis]
	}

	// Untuk surat lain (Domisili, Usaha, SKTM Umum),
	// HANYA tanya field yang TIDAK ADA di BaseFields (data NIK)
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
		// Jika field TIDAK DITEMUKAN di BaseFields, maka kita perlu menanyakannya
		if !isBase {
			neededFields = append(neededFields, field)
		}
	}
	return neededFields
}

func BuildDataMap(data db.DataPenduduk) map[string]string {
	// --- Persiapan Data Dasar ---
	
	// 1. Gabungkan Tempat & Tanggal Lahir
	ttl := fmt.Sprintf("%s, %s", data.TempatLahir.String, db.FormatDate(data.TanggalLahir))
	
	// 2. Tentukan Jenis Kelamin
	jk := "Laki-laki"
	if data.SexID.Int64 == 2 {
		jk = "Perempuan"
	}
	
	// 3. Buat Alamat Lengkap
	alamatLengkap := fmt.Sprintf("%s, Dusun %s, RT/RW %s/%s", data.Alamat.String, data.Dusun.String, data.RT.String, data.RW.String)

	// --- Pemetaan Data ke Placeholder ---
	// Peta ini akan mengisi semua placeholder yang diketahui dari NIK.
	// Field yang tidak ada di sini (seperti ALASANPERLU, NAMA.C) 
	// akan ditanyakan secara manual oleh bot.
	return map[string]string{
		
		// --- Data Dasar (Auto-filled dari NIK pengguna) ---
		"NAMA":      data.Nama.String,
		"TTL":       ttl,
		"JK":        jk,
		"AGAMA":     data.AgamaNama.String, // (Membutuhkan JOIN di GetDataPendudukByNIK)
		"NIK":       data.NIK.String,
		"PEKERJAAN": data.PekerjaanNama.String, // (Membutuhkan JOIN di GetDataPendudukByNIK)
		"ALAMAT":    alamatLengkap,
		"TANGGAL":   time.Now().Format("02 January 2006"),


		// Untuk SK Usaha & SKTM Umum (TTLnU = TTL Pemohon)
		"TTLnU":     ttl, 
		
		// Untuk SKTM Tanggungan (Otomatis mengisi data Orang Tua/Wali)
		// Kita asumsikan "Orang Tua" adalah Ayah
		"NAMA.P":      data.NamaAyah.String,
		"NIK.P":       data.NikAyah.String,
		"PEKERJAAN.P": "", // Kita tidak tahu pekerjaan Ayah dari data NIK anak
		"AGAMA.P":     "", // Kita tidak tahu agama Ayah
		"TTL.P":       "", // Kita tidak tahu TTL Ayah
		"JK.P":        "Laki-laki", // Asumsi Ayah
		"STATUS.P":    "Ayah Kandung", // Asumsi
		"ALAMAT.P":    alamatLengkap, // Asumsi alamat orang tua = alamat anak
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
