package surat

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"go.mau.fi/whatsmeow"
)

const (
	STEP_SURAT_MENU  = 500
	STEP_SURAT_INPUT = 501
)

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

func Handle(dbConn *sqlx.DB, jid, text string, session *db.DataEntrySession, client *whatsmeow.Client) []string {
	text = strings.TrimSpace(text)

	switch session.CurrentStep {
	case 0:
		_ = db.UpdateStepOnly(dbConn, jid, STEP_SURAT_MENU)
		if text == "init" {
			return []string{
				"*Pilih Jenis Surat:*\n" +
					"1. Surat Domisili\n" +
					"2. Surat Usaha\n" +
					"3. SKTM Umum\n" +
					"4. SKTM Tanggungan\n" +
					"5. Surat Kematian",
			}
		}

	case STEP_SURAT_MENU:
		return handleJenisSurat(dbConn, jid, text)

	case STEP_SURAT_INPUT:
		return handleInputSurat(dbConn, jid, text, client)

	default:
		return []string{"Perintah tidak dikenali. Ketik 2 untuk membuat surat."}
	}
	return []string{"Perintah tidak dikenali. Ketik 2 untuk membuat surat."}
}

func handleJenisSurat(dbConn *sqlx.DB, jid, text string) []string {
	suratMap := map[string]JenisSurat{
		"1": DOMISILI,
		"2": USAHA,
		"3": SKTM_UMUM,
		"4": SKTM_TANGGUNGAN,
		"5": KEMATIAN,
	}
	jenis, ok := suratMap[text]
	if !ok {
		return []string{"Masukkan angka 1–5 sesuai jenis surat."}
	}

	fieldList := append(BaseFields, SuratFields[jenis]...)
	fieldStr := strings.Join(fieldList, ",")

	if err := db.SetEditField(dbConn, jid, string(jenis)); err != nil {
		log.Printf("[SURAT-DB] Gagal menyimpan jenis surat: %v", err)
	}
	if err := db.UpdateStepOnly(dbConn, jid, STEP_SURAT_INPUT); err != nil {
		log.Printf("[SURAT-DB] Gagal update step surat: %v", err)
	}

	sessionKey := fmt.Sprintf("surat_fields_%s", jid)
	_ = db.SaveTemporary(sessionKey, fieldStr)
	_ = db.SaveTemporary("surat_field_now_"+jid, fieldList[0])

	return []string{getPrompt(fieldList[0])}
}

func handleInputSurat(dbConn *sqlx.DB, jid, text string, client *whatsmeow.Client) []string {
	currentField, _ := db.LoadTemporary("surat_field_now_" + jid)
	fieldListStr, _ := db.LoadTemporary("surat_fields_" + jid)
	fieldList := strings.Split(fieldListStr, ",")

	_ = db.SaveTemporary(jid+"_field_"+currentField, text)

	next := nextField(fieldList, currentField)
	if next != "" {
		_ = db.SaveTemporary("surat_field_now_"+jid, next)
		return []string{getPrompt(next)}
	}

	data := make(map[string]string)
	for _, f := range fieldList {
		val, _ := db.LoadTemporary(jid + "_field_" + strings.TrimSpace(f))
		data[strings.TrimSpace(f)] = val
	}
	data["TANGGAL"] = time.Now().Format("02 January 2006")

	jenisStr, _ := db.GetEditField(dbConn, jid)
	jenis := JenisSurat(jenisStr)

	path, err := GenerateAsync(jenis, data, "temp", jid, client)
	if err != nil {
		log.Printf("[SURAT-ERROR] %v", err)
		return []string{"Terjadi kesalahan saat membuat surat."}
	}

	db.DeleteDataEntrySession(dbConn, jid)
	db.ClearTemporaryByPrefix(jid + "_")

	return []string{
		fmt.Sprintf("Surat *%s* sedang diproses...\nFile LaTeX: %s", jenis, path),
	}
}

func nextField(fields []string, current string) string {
	for i, f := range fields {
		if strings.TrimSpace(f) == strings.TrimSpace(current) && i+1 < len(fields) {
			return strings.TrimSpace(fields[i+1])
		}
	}
	return ""
}

func getPrompt(field string) string {
	if prompt, ok := FieldPrompts[field]; ok {
		return prompt
	}
	return fmt.Sprintf("Masukkan data untuk *%s*:", field)
}
