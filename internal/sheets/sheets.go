package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db"
	// "github.com/sinavarasina/SAS-BOT/internal/surat" // <-- PERBAIKAN: HAPUS IMPORT INI
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient sekarang memegang SEMUA ID Spreadsheet
type SheetsClient struct {
	Service                *sheets.Service
	DataSpreadsheetID      string // Untuk Data Diri
	PengaduanSpreadsheetID string // Untuk Pengaduan
	SuratSpreadsheetID     string
	UlasanSpreadsheetID    string
}

func InitSheetsClient() (*SheetsClient, error) {
	ctx := context.Background()
	credentialFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialFile == "" {
		log.Fatal("Environment variable GOOGLE_APPLICATION_CREDENTIALS tidak di-set")
	}
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialFile), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}
	
	dataSpreadsheetID := os.Getenv("DATA_SPREADSHEET_ID")
	if dataSpreadsheetID == "" { log.Fatal("DATA_SPREADSHEET_ID tidak di-set") }
	
	pengaduanSpreadsheetID := os.Getenv("PENGADUAN_SPREADSHEET_ID")
	if pengaduanSpreadsheetID == "" { log.Println("Peringatan: PENGADUAN_SPREADSHEET_ID tidak di-set.") }
	
	suratSpreadsheetID := os.Getenv("SURAT_SPREADSHEET_ID")
	if suratSpreadsheetID == "" { log.Fatal("SURAT_SPREADSHEET_ID tidak di-set") }

	ulasanSpreadsheetID := os.Getenv("ULASAN_SPREADSHEET_ID")
	if ulasanSpreadsheetID == "" { log.Fatal("ULASAN_SPREADSHEET_ID tidak di-set") }

	return &SheetsClient{
		Service:                srv,
		DataSpreadsheetID:      dataSpreadsheetID,
		PengaduanSpreadsheetID: pengaduanSpreadsheetID,
		SuratSpreadsheetID:     suratSpreadsheetID,
		UlasanSpreadsheetID:    ulasanSpreadsheetID,
	}, nil
}

// --- FUNGSI DATA PENDUDUK (Dipersingkat 19 langkah/21 kolom) ---
func (c *SheetsClient) buildRowDataPenduduk(s db.DataEntrySession) []interface{} {
	return []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), s.JID,
		s.Dusun.String, s.RT.String, s.Nama.String, s.NoKK.String, s.NIK.String,
		s.SexNama.String, s.TempatLahir.String, db.FormatDate(s.TanggalLahir),
		s.AgamaNama.String, s.PendidikanKKNama.String, s.PendidikanSedangNama.String,
		s.PekerjaanNama.String, s.StatusKawinNama.String, s.KKLevelNama.String,
		s.WarganegaraNama.String, s.NamaAyah.String, s.NamaIbu.String,
		s.StatusDasarNama.String, s.SukuNama.String,
	}
}
func (c *SheetsClient) AppendDataPenduduk(s db.DataEntrySession) {
	rangeData := "Data_penduduk!A2"
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowDataPenduduk(s))
	_, err := c.Service.Spreadsheets.Values.Append(c.DataSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal AppendDataPenduduk ke Google Sheet: %v", err)
	} else {
		log.Println("Berhasil AppendDataPenduduk ke Google Sheet.")
	}
}
func (c *SheetsClient) FindRowByNIK(nik string) (int, error) {
	readRange := "Data_penduduk!G:G" // NIK di Kolom G
	resp, err := c.Service.Spreadsheets.Values.Get(c.DataSpreadsheetID, readRange).Do()
	if err != nil { return 0, fmt.Errorf("gagal membaca Sheet: %w", err) }
	if len(resp.Values) == 0 { return 0, fmt.Errorf("NIK tidak ditemukan (Sheet kosong)") }
	for i, row := range resp.Values {
		if len(row) > 0 {
			if cellValue, ok := row[0].(string); ok && cellValue == nik {
				return i + 1, nil
			}
		}
	}
	return 0, fmt.Errorf("NIK tidak ditemukan")
}
func (c *SheetsClient) UpdateRowData(rowNum int, s db.DataEntrySession) error {
	rangeData := fmt.Sprintf("Data_penduduk!A%d:U%d", rowNum, rowNum) // 21 kolom (A-U)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowDataPenduduk(s))
	_, err := c.Service.Spreadsheets.Values.Update(c.DataSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal meng-update Google Sheet: %v", err)
		return err
	}
	log.Println("Berhasil meng-update Google Sheet.")
	return nil
}
func (c *SheetsClient) DeleteRowDataPenduduk(rowNum int) error {
	sheetID, err := c.findSheetIdByName(c.DataSpreadsheetID, "Data_penduduk")
	if err != nil { return err }
	req := sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "ROWS",
				StartIndex: int64(rowNum - 1),
				EndIndex:   int64(rowNum),
			},
		},
	}
	batchReq := sheets.BatchUpdateSpreadsheetRequest{ Requests: []*sheets.Request{&req} }
	_, err = c.Service.Spreadsheets.BatchUpdate(c.DataSpreadsheetID, &batchReq).Do()
	if err != nil {
		log.Printf("Gagal menghapus baris dari Google Sheet: %v", err)
		return err
	}
	log.Println("Berhasil menghapus baris dari Google Sheet.")
	return nil
}

// --- FUNGSI PENGADUAN (Tidak berubah) ---
func (c *SheetsClient) AppendPengaduan(aduan db.Pengaduan, publicID string) {
	rangeData := "Sheet1!A2" // Ganti 'Sheet1' dengan nama tab pengaduan Anda
	gambarFormula := fmt.Sprintf("=IMAGE(\"%s\")", aduan.PictPath)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		time.Now().Format("2006-01-02 15:04:05"),
		aduan.UserJID, aduan.Deskripsi, aduan.PictPath,
		gambarFormula, "Belum Diproses", publicID,
	})
	_, err := c.Service.Spreadsheets.Values.Append(c.PengaduanSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[ERROR] Gagal menulis Pengaduan ke Google Sheet: %v", err)
	} else {
		log.Println("[INFO] Berhasil menulis Pengaduan ke Google Sheet.")
	}
}
func (c *SheetsClient) GetPengaduanStatus(publicID string) (string, error) {
	readRange := "Sheet1!F:G" // Kolom F (Status) dan G (ID)
	resp, err := c.Service.Spreadsheets.Values.Get(c.PengaduanSpreadsheetID, readRange).Do()
	if err != nil { return "", fmt.Errorf("gagal membaca Sheet Pengaduan: %w", err) }
	if len(resp.Values) == 0 { return "", fmt.Errorf("Sheet Pengaduan kosong") }
	for _, row := range resp.Values {
		if len(row) >= 2 {
			statusCellValue, statusOk := row[0].(string)
			idCellValue, idOk := row[1].(string)
			if idOk && statusOk && idCellValue == publicID {
				return statusCellValue, nil
			}
		}
	}
	return "", fmt.Errorf("ID Pengaduan tidak ditemukan")
}


// --- FUNGSI SURAT (DIPERBARUI) ---

// PERBAIKAN: Tanda tangan fungsi diubah dari 'surat.JenisSurat' ke 'string'
func (c *SheetsClient) AppendSuratLog(jenisSuratStr string, nama, unikID, tgl, status, fileURL string) {
	
	var sheetName string
	switch jenisSuratStr {
	case "sk_domisili.tex":
		sheetName = "SK_Domisili"
	case "sk_usaha.tex":
		sheetName = "SK_Usaha"
	case "sktm_umum.tex":
		sheetName = "SKTM_Umum"
	case "sktm_tanggungan.tex":
		sheetName = "SKTM_Tanggungan"
	case "sk_kematian.tex":
		sheetName = "SK_Kematian"
	default:
		log.Printf("[SHEETS-ERROR] Nama sheet tidak diketahui untuk jenis surat: %s", jenisSuratStr)
		return
	}
	
	rangeData := fmt.Sprintf("%s!A2", sheetName)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		tgl, nama, unikID, status, fileURL,
	})

	_, err := c.Service.Spreadsheets.Values.Append(c.SuratSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[SHEETS-ERROR] Gagal AppendSuratLog ke %s: %v", sheetName, err)
	}
}

func (c *SheetsClient) GetSuratStatus(unikID string) (string, error) {
	sheetNames := []string{"SK_Domisili", "SK_Usaha", "SKTM_Umum", "SKTM_Tanggungan", "SK_Kematian"}
	
	for _, sheetName := range sheetNames {
		readRange := fmt.Sprintf("%s!C:D", sheetName) // Kolom C (ID), Kolom D (Status)
		
		resp, err := c.Service.Spreadsheets.Values.Get(c.SuratSpreadsheetID, readRange).Do()
		if err != nil {
			log.Printf("[SHEETS-WARN] Gagal membaca tab %s: %v", sheetName, err)
			continue
		}

		for _, row := range resp.Values {
			if len(row) >= 2 {
				idCellValue, idOk := row[0].(string)
				statusCellValue, statusOk := row[1].(string)
				if idOk && statusOk && strings.EqualFold(idCellValue, unikID) {
					return statusCellValue, nil
				}
			}
		}
	}
	return "", fmt.Errorf("ID Surat tidak ditemukan")
}


func (c *SheetsClient) AppendUlasan(sheetName, layanan, rating, jid string) {
	rangeData := fmt.Sprintf("%s!A2", sheetName)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // Timestamp
		jid,
		layanan,
		rating,
	})

	_, err := c.Service.Spreadsheets.Values.Append(c.UlasanSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[ULASAN-ERROR] Gagal AppendUlasan ke %s: %v", sheetName, err)
	} else {
		log.Printf("[ULASAN] Berhasil menyimpan ulasan ke %s", sheetName)
	}
}


// --- FUNGSI HELPER (findSheetIdByName) ---
func (c *SheetsClient) findSheetIdByName(spreadsheetID, sheetName string) (int64, error) {
	resp, err := c.Service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, err
	}
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("Sheet dengan nama '%s' tidak ditemukan", sheetName)
}
