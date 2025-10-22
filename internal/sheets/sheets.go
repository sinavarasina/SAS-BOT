// file: internal/sheets/sheets.go

package sheets

import (
	"context"
	"log"
	"os"
	"time"
	"fmt"

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
	
	// Nama file kredensial diambil dari environment variable
	credentialFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialFile == "" {
		log.Fatal("Environment variable GOOGLE_APPLICATION_CREDENTIALS tidak di-set")
	}

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		return nil, err
	}
	
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		log.Fatal("Environment variable SPREADSHEET_ID tidak di-set")
	}

	return &SheetsClient{
		Service:       srv,
		SpreadsheetID: spreadsheetID,
	}, nil
}

// WritePengaduan menulis satu baris data pengaduan ke spreadsheet.
func (c *SheetsClient) WritePengaduan(aduan db.Pengaduan) {
	rangeData := "A2" // Menulis ke sheet pertama, mulai dari sel A1 (akan di-append).
	gambarFormula := fmt.Sprintf("=IMAGE(\"%s\")", aduan.PictPath)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // Timestamp
		aduan.UserJID,
		aduan.Deskripsi,
		aduan.PictPath,
		gambarFormula,
	})

	_, err := c.Service.Spreadsheets.Values.Append(c.SpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Gagal menulis ke Google Sheet: %v", err)
	} else {
		log.Println("Berhasil menulis pengaduan ke Google Sheet.")
	}
}
