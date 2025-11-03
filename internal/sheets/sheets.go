
package sheets

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sinavarasina/SAS-BOT/internal/db"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient adalah struct untuk menyimpan service Google Sheets.
type SheetsClient struct {
	Service       *sheets.Service
	SpreadsheetID string
}

// InitSheetsClient membuat koneksi ke Google Sheets API.
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
	
	spreadsheetID := os.Getenv("DATA_SPREADSHEET_ID")
	if spreadsheetID == "" {
		log.Fatal("Environment variable SPREADSHEET_ID tidak di-set")
	}

	return &SheetsClient{
		Service:       srv,
		SpreadsheetID: spreadsheetID,
	}, nil
}

// WriteDataPenduduk menulis satu baris data penduduk lengkap ke spreadsheet.
func (c *SheetsClient) WriteDataPenduduk(s db.DataEntrySession) {
	// Menulis ke Sheet pertama (A1)
	rangeData := "Data_penduduk!A2" 

	var vr sheets.ValueRange
	
	// Siapkan 1 baris data (42 kolom)
	// Kita gunakan field Nama (misal SexNama) agar mudah dibaca di sheet, bukan ID-nya
	rowData := []interface{}{
		s.Alamat.String,
		s.Dusun.String,
		s.RW.String,
		s.RT.String,
		s.Nama.String,
		s.NoKK.String,
		s.NIK.String,
		s.SexNama.String, 			// Menggunakan ...Nama
		s.TempatLahir.String,
		db.FormatDate(s.TanggalLahir), // Menggunakan helper formatDate
		s.AgamaNama.String,
		s.PendidikanKKNama.String,
		s.PendidikanSedangNama.String,
		s.PekerjaanNama.String,
		s.StatusKawinNama.String,
		s.KKLevelNama.String,
		s.WarganegaraNama.String,
		s.NikAyah.String,
		s.NamaAyah.String,
		s.NikIbu.String,
		s.NamaIbu.String,
		s.GolonganDarahNama.String,
		s.AktaLahir.String,
		s.DokumenPassport.String,
		db.FormatDate(s.TanggalAkhirPassport),
		s.DokumenKitas.String,
		s.AktaPerkawinan.String,
		db.FormatDate(s.TanggalPerkawinan),
		s.AktaPerceraian.String,
		db.FormatDate(s.TanggalPerceraian),
		s.CacatNama.String,
		s.CaraKBNama.String,
		s.HamilNama.String,
		s.KTPElNama.String,
		s.StatusRekamNama.String,
		s.AlamatSekarang.String,
		s.StatusDasarNama.String,
		s.SukuNama.String,
		s.TagCard.String,
		s.AsuransiNama.String,
		s.NoAsuransi.String,
	}
	
	vr.Values = append(vr.Values, rowData)

	_, err := c.Service.Spreadsheets.Values.Append(c.SpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal menulis Data Penduduk ke Google Sheet: %v", err)
	} else {
		log.Println("Berhasil menulis Data Penduduk ke Google Sheet.")
	}
}

func (c *SheetsClient) CheckNIKExistsInSheet(nik string) (bool, error) {
	// Membaca "Sheet1!G:G" -> Sheet1, Kolom G, semua baris.
	readRange := "Data_penduduk!G:G"

	resp, err := c.Service.Spreadsheets.Values.Get(c.SpreadsheetID, readRange).Do()
	if err != nil {
		log.Printf("Gagal membaca Google Sheet: %v", err)
		return false, fmt.Errorf("gagal memverifikasi NIK di Sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		log.Println("Kolom NIK (G) di Sheet kosong, tidak ada duplikat.")
		return false, nil // Tidak ada data, berarti tidak duplikat
	}

	// Loop setiap baris di kolom G
	for _, row := range resp.Values {
		if len(row) > 0 {
			// Cek apakah nilai sel di kolom G == NIK yang dicari
			if cellValue, ok := row[0].(string); ok && cellValue == nik {
				log.Printf("NIK duplikat ditemukan di Google Sheet: %s", nik)
				return true, nil // Ditemukan
			}
		}
	}

	return false, nil // Tidak ditemukan
}
