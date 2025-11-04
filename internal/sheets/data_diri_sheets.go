
package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient adalah struct untuk menyimpan service Google Sheets.
type Data_SheetsClient struct {
	Service       *sheets.Service
	SpreadsheetID string
}

// InitSheetsClient membuat koneksi ke Google Sheets API.
func InitSheetsClient() (*Data_SheetsClient, error) {
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

	return &Data_SheetsClient{
		Service:       srv,
		SpreadsheetID: spreadsheetID,
	}, nil
}

func (c *Data_SheetsClient) buildRowData(s db.DataEntrySession) []interface{} {
	return []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // A
		s.Alamat.String,                 // B 
		s.Dusun.String,                  // C 
		s.RW.String,                     // D 
		s.RT.String,                     // E 
		s.Nama.String,                   // F 
		s.NoKK.String,                   // G 
		s.NIK.String,                    // H  
		s.SexNama.String,                // I 
		s.TempatLahir.String,             // J 
		db.FormatDate(s.TanggalLahir),   // K 
		s.AgamaNama.String,              // L 
		s.PendidikanKKNama.String,       // M 
		s.PendidikanSedangNama.String,   // N 
		s.PekerjaanNama.String,          // O 
		s.StatusKawinNama.String,        // P 
		s.KKLevelNama.String,            // Q 
		s.WarganegaraNama.String,        // R 
		s.NikAyah.String,                // S 
		s.NamaAyah.String,               // T 
		s.NikIbu.String,                 // U 
		s.NamaIbu.String,                // V 
		s.GolonganDarahNama.String,      // W 
		s.AktaLahir.String,              // X 
		s.DokumenPassport.String,        // Y 
		db.FormatDate(s.TanggalAkhirPassport), // Z 
		s.DokumenKitas.String,           // AA 
		s.AktaPerkawinan.String,         // AB 
		db.FormatDate(s.TanggalPerkawinan), // AC 
		s.AktaPerceraian.String,         // AD 
		db.FormatDate(s.TanggalPerceraian), // AE 
		s.CacatNama.String,              // AF 
		s.CaraKBNama.String,             // AG 
		s.HamilNama.String,              // AH 
		s.KTPElNama.String,              // AI 
		s.StatusRekamNama.String,        // AJ 
		s.AlamatSekarang.String,         // AK 
		s.StatusDasarNama.String,        // AL 
		s.SukuNama.String,               
		s.TagCard.String,                
		s.AsuransiNama.String,           
		s.NoAsuransi.String,             
	}
}

func (c *Data_SheetsClient) AppendDataPenduduk(s db.DataEntrySession) {
	rangeData := "Data_penduduk!A2"
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowData(s))

	_, err := c.Service.Spreadsheets.Values.Append(c.SpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal AppendDataPenduduk ke Google Sheet: %v", err)
	} else {
		log.Println("Berhasil AppendDataPenduduk ke Google Sheet.")
	}
}


func (c *Data_SheetsClient) FindRowByNIK(nik string) (int, error) {
	readRange := "Data_penduduk!H:H" // NIK di Kolom H 

	resp, err := c.Service.Spreadsheets.Values.Get(c.SpreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("gagal membaca Sheet: %w", err)
	}
	if len(resp.Values) == 0 {
		return 0, fmt.Errorf("NIK tidak ditemukan (Sheet kosong)")
	}

	for i, row := range resp.Values {
		if len(row) > 0 {
			if cellValue, ok := row[0].(string); ok && cellValue == nik {
				return i + 1, nil // Kembalikan nomor baris (1-based index)
			}
		}
	}
	return 0, fmt.Errorf("NIK tidak ditemukan")
}

func (c *Data_SheetsClient) UpdateRowData(rowNum int, s db.DataEntrySession) error {
	rangeData := fmt.Sprintf("Data_penduduk!A%d:AQ%d", rowNum, rowNum) // 43 Kolom
	
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowData(s))

	_, err := c.Service.Spreadsheets.Values.Update(c.SpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal meng-update Google Sheet: %v", err)
		return err
	}
	log.Println("Berhasil meng-update Google Sheet.")
	return nil
}

func (c *Data_SheetsClient) findSheetIdByName(sheetName string) (int64, error) {
	resp, err := c.Service.Spreadsheets.Get(c.SpreadsheetID).Do()
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

