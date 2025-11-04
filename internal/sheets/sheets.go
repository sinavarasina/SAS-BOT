
package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db" // Pastikan db.Pengaduan ada jika Anda menggabungkan
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient sekarang memegang SEMUA ID Spreadsheet
type SheetsClient struct {
	Service               *sheets.Service
	DataSpreadsheetID     string // Untuk Data Diri
	PengaduanSpreadsheetID string // Untuk Pengaduan
}

// InitSheetsClient sekarang membaca KEDUA ID dari .env
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
	
	// Baca ID Data Diri
	dataSpreadsheetID := os.Getenv("DATA_SPREADSHEET_ID")
	if dataSpreadsheetID == "" {
		log.Fatal("Environment variable DATA_SPREADSHEET_ID tidak di-set")
	}
	
	// Baca ID Pengaduan
	pengaduanSpreadsheetID := os.Getenv("PENGADUAN_SPREADSHEET_ID")
	if pengaduanSpreadsheetID == "" {
		log.Println("Peringatan: PENGADUAN_SPREADSHEET_ID tidak di-set. Fitur pengaduan tidak akan bekerja.")
		// Jangan Fatal, agar bot tetap jalan
	}

	return &SheetsClient{
		Service:               srv,
		DataSpreadsheetID:     dataSpreadsheetID,
		PengaduanSpreadsheetID: pengaduanSpreadsheetID,
	}, nil
}


// buildRowDataPenduduk adalah helper untuk data penduduk
func (c *SheetsClient) buildRowDataPenduduk(s db.DataEntrySession) []interface{} {
	// (Menggunakan 43 kolom dari file Anda sebelumnya)
	return []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // A
		s.Alamat.String,                 // C
		s.Dusun.String,                  // D
		s.RW.String,                     // E
		s.RT.String,                     // F
		s.Nama.String,                   // G
		s.NoKK.String,                   // H
		s.NIK.String,                    // I
		s.SexNama.String,                // J
		s.TempatLahir.String,             // K
		db.FormatDate(s.TanggalLahir),   // L
		s.AgamaNama.String,              // M
		s.PendidikanKKNama.String,       // N
		s.PendidikanSedangNama.String,   // O
		s.PekerjaanNama.String,          // P
		s.StatusKawinNama.String,        // Q
		s.KKLevelNama.String,            // R
		s.WarganegaraNama.String,        // S
		s.NikAyah.String,                // T
		s.NamaAyah.String,               // U
		s.NikIbu.String,                 // V
		s.NamaIbu.String,                // W
		s.GolonganDarahNama.String,      // X
		s.AktaLahir.String,              // Y
		s.DokumenPassport.String,        // Z
		db.FormatDate(s.TanggalAkhirPassport), // AA
		s.DokumenKitas.String,           // AB
		s.AktaPerkawinan.String,         // AC
		db.FormatDate(s.TanggalPerkawinan), // AD
		s.AktaPerceraian.String,         // AE
		db.FormatDate(s.TanggalPerceraian), // AF
		s.CacatNama.String,              // AG
		s.CaraKBNama.String,             // AH
		s.HamilNama.String,              // AI
		s.KTPElNama.String,              // AJ
		s.StatusRekamNama.String,        // AK
		s.AlamatSekarang.String,         // AL
		s.StatusDasarNama.String,        // AM
		s.SukuNama.String,               // AN
		s.TagCard.String,                // AO
		s.AsuransiNama.String,           // AP
		s.NoAsuransi.String,             // AQ
	}
}

// AppendDataPenduduk menulis ke Sheet Data Diri
func (c *SheetsClient) AppendDataPenduduk(s db.DataEntrySession) {
	rangeData := "Data_penduduk!A2" // Asumsi nama tab 'Data_penduduk'
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowDataPenduduk(s))

	// Menggunakan c.DataSpreadsheetID
	_, err := c.Service.Spreadsheets.Values.Append(c.DataSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal AppendDataPenduduk ke Google Sheet: %v", err)
	} else {
		log.Println("Berhasil AppendDataPenduduk ke Google Sheet.")
	}
}

// FindRowByNIK mencari NIK di Sheet Data Diri
func (c *SheetsClient) FindRowByNIK(nik string) (int, error) {
	readRange := "Data_penduduk!H:H" // NIK di Kolom H 

	// Menggunakan c.DataSpreadsheetID
	resp, err := c.Service.Spreadsheets.Values.Get(c.DataSpreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("gagal membaca Sheet: %w", err)
	}
	if len(resp.Values) == 0 {
		return 0, fmt.Errorf("NIK tidak ditemukan (Sheet kosong)")
	}

	for i, row := range resp.Values {
		if len(row) > 0 {
			if cellValue, ok := row[0].(string); ok && cellValue == nik {
				return i + 1, nil 
			}
		}
	}
	return 0, fmt.Errorf("NIK tidak ditemukan")
}

// UpdateRowData menimpa data di Sheet Data Diri
func (c *SheetsClient) UpdateRowData(rowNum int, s db.DataEntrySession) error {
	rangeData := fmt.Sprintf("Data_penduduk!A%d:AQ%d", rowNum, rowNum)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, c.buildRowDataPenduduk(s))

	// Menggunakan c.DataSpreadsheetID
	_, err := c.Service.Spreadsheets.Values.Update(c.DataSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal meng-update Google Sheet: %v", err)
		return err
	}
	log.Println("Berhasil meng-update Google Sheet.")
	return nil
}

// --- FUNGSI UNTUK PENGADUAN (JIKA ANDA GABUNGKAN) ---

// (Ini adalah fungsi dari file sheets.go Anda yang lama, sekarang digabung di sini)
// Anda mungkin perlu membuat struct db.Pengaduan

func (c *SheetsClient) AppendPengaduan(aduan db.Pengaduan) {
	rangeData := "Sheet1!A2" // Ganti nama tab jika perlu
	gambarFormula := fmt.Sprintf("=IMAGE(\"%s\")", aduan.PictPath)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		time.Now().Format("2006-01-02 15:04:05"),
		aduan.UserJID,
		aduan.Deskripsi,
		aduan.PictPath,
		gambarFormula,
		"Belum Diproses",
	})

	// Menggunakan c.PengaduanSpreadsheetID
	_, err := c.Service.Spreadsheets.Values.Append(c.PengaduanSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[ERROR] Gagal menulis Pengaduan ke Google Sheet: %v", err)
	} else {
		log.Println("[INFO] Berhasil menulis Pengaduan ke Google Sheet.")
	}
}

// --- FUNGSI HELPER (findSheetIdByName, DeleteRow) ---
// (Fungsi-fungsi ini perlu tahu ID spreadsheet mana yang harus digunakan)

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

