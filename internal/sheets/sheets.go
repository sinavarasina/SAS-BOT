package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sinavarasina/SAS-BOT/internal/db"
	// "github.com/sinavarasina/SAS-BOT/internal/surat"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient sekarang memegang SEMUA ID Spreadsheet
type SheetsClient struct {
	Service               *sheets.Service
	DataSpreadsheetID     string // Untuk Data Diri
	PengaduanSpreadsheetID string // Untuk Pengaduan
	SuratSpreadsheetID     string
	UlasanSpreadsheetID    string
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
	
	// baca id surat
	suratSpreadsheetID := os.Getenv("SURAT_SPREADSHEET_ID")
	if suratSpreadsheetID == "" {
		log.Println("Environment variable SURAT_SPREADSHEET_ID tidak di-set")
	}

	//baca id feedback
	ulasanSpreadsheetID := os.Getenv("ULASAN_SPREADSHEET_ID")
	if ulasanSpreadsheetID == "" {
		log.Fatal("Environment variable ULASAN_SPREADSHEET_ID tidak di-set")
	}

	return &SheetsClient{
		Service:               srv,
		DataSpreadsheetID:     dataSpreadsheetID,
		PengaduanSpreadsheetID: pengaduanSpreadsheetID,
		SuratSpreadsheetID: suratSpreadsheetID,
		UlasanSpreadsheetID:    ulasanSpreadsheetID,
	}, nil
}


// buildRowDataPenduduk adalah helper untuk data penduduk
func (c *SheetsClient) buildRowDataPenduduk(s db.DataEntrySession) []interface{} {
	// 39 kolom (Dusun, RT, Nama, No KK, NIK, Sex, Tempat Lahir, Tanggal Lahir, Agama, Pendidikan KK, 
	// Pendidikan Sedang, Pekerjaan, Status Kawin, Level KK, Warganegara, Nama Ayah, Nama Ibu,
	// Status Dasar, Suku, NIK Ayah, NIK Ibu, Golongan Darah, Akta Lahir, Dokumen Passport,
	// Tanggal Akhir Passport, Dokumen KITAS, Akta Perkawinan, Tanggal Perkawinan, Akta Perceraian,
	// Tanggal Perceraian, Cacat, Cara KB, Hamil, KTP El, Status Rekam, Alamat Sekarang, Tag Card, Asuransi, No Asuransi)
	return []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // A (Timestamp)
		s.Dusun.String,                  // B
		s.RT.String,                     // C
		s.Nama.String,                   // D
		s.NoKK.String,                   // E
		s.NIK.String,                    // F
		s.SexNama.String,                // G
		s.TempatLahir.String,            // H
		db.FormatDate(s.TanggalLahir),   // I
		s.AgamaNama.String,              // J
		s.PendidikanKKNama.String,       // K
		s.PendidikanSedangNama.String,   // L
		s.PekerjaanNama.String,          // M
		s.StatusKawinNama.String,        // N
		s.KKLevelNama.String,            // O
		s.WarganegaraNama.String,        // P
		s.NamaAyah.String,               // Q
		s.NamaIbu.String,                // R
		s.StatusDasarNama.String,        // S
		s.SukuNama.String,               // T
		s.NikAyah.String,                // U
		s.NikIbu.String,                 // V
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
		s.TagCard.String,                // AL
		s.AsuransiNama.String,           // AM
		s.NoAsuransi.String,             // AN
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
	readRange := "Data_penduduk!F:F" // NIK di Kolom F (berubah dari H)

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
	rangeData := fmt.Sprintf("Data_penduduk!A%d:AN%d", rowNum, rowNum) // Update dari AQ ke AN (39 kolom)
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

func (c *SheetsClient) AppendPengaduan(aduan db.Pengaduan, publicID string) {
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
		publicID,
	})

	// Menggunakan c.PengaduanSpreadsheetID:
	_, err := c.Service.Spreadsheets.Values.Append(c.PengaduanSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[ERROR] Gagal menulis Pengaduan ke Google Sheet: %v", err)
	} else {
		log.Println("[INFO] Berhasil menulis Pengaduan ke Google Sheet.")
	}
}

func (c *SheetsClient) AppendUlasan(sheetName, layanan, rating, jid string) {
	rangeData := fmt.Sprintf("%s!A2", sheetName) // Selalu append
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		time.Now().Format("2006-01-02 15:04:05"), // Timestamp
		jid,     // JID Pengguna
		layanan, // Misal: "SK Domisili" atau "Input Data Diri"
		rating,  // Misal: "5"
	})

	// Gunakan c.UlasanSpreadsheetID
	_, err := c.Service.Spreadsheets.Values.Append(c.UlasanSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[ULASAN-ERROR] Gagal AppendUlasan ke %s: %v", sheetName, err)
	} else {
		log.Printf("[ULASAN] Berhasil menyimpan ulasan ke %s", sheetName)
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

func (c *SheetsClient) GetPengaduanStatus(publicID string) (string, error) {
	// 1. Tentukan rentang kolom yang akan dibaca
	//    Kita perlu Kolom F (Status) dan Kolom G (PublicID)
	readRange := "Sheet1!F:G"

	resp, err := c.Service.Spreadsheets.Values.Get(c.PengaduanSpreadsheetID, readRange).Do()
	if err != nil {
		return "", fmt.Errorf("gagal membaca Sheet Pengaduan: %w", err)
	}

	if len(resp.Values) == 0 {
		return "", fmt.Errorf("Sheet Pengaduan kosong")
	}

	// 2. Loop melalui setiap baris yang didapat
	for _, row := range resp.Values {
		// Pastikan baris memiliki cukup kolom (minimal 2 untuk F dan G)
		if len(row) >= 2 {
			// Kolom G (index 1) adalah PublicID
			idCellValue, idOk := row[1].(string)
			// Kolom F (index 0) adalah Status
			statusCellValue, statusOk := row[0].(string)

			if idOk && statusOk && idCellValue == publicID {
				// Ditemukan!
				return statusCellValue, nil
			}
		}
	}

	// 3. Jika loop selesai tanpa menemukan ID
	return "", fmt.Errorf("ID Pengaduan tidak ditemukan")
}

// AppendSuratLog mencatat pengajuan surat ke tab yang sesuai
func (c *SheetsClient) AppendSuratLog(jenisSuratStr string, nama, unikID, tgl, status, fileURL string) {
	
	// Tentukan nama Tab (Sheet) berdasarkan jenis surat
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
	
	rangeData := fmt.Sprintf("%s!A2", sheetName) // Selalu append
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		tgl,       // Tanggal Pengajuan
		nama,      // Nama Pengaju
		unikID,    // No Unik Surat
		status,    // Status
		fileURL,   // Link ke PDF
	})

	_, err := c.Service.Spreadsheets.Values.Append(c.SuratSpreadsheetID, rangeData, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("[SHEETS-ERROR] Gagal AppendSuratLog ke %s: %v", sheetName, err)
	}
}

// GetSuratStatus mencari ID unik di SEMUA tab surat
func (c *SheetsClient) GetSuratStatus(unikID string) (string, error) {
	sheetNames := []string{"SK_Domisili", "SK_Usaha", "SKTM_Umum", "SKTM_Tanggungan", "SK_Kematian"}
	
	for _, sheetName := range sheetNames {
		// Asumsi ID Unik ada di Kolom C dan Status di Kolom D
		readRange := fmt.Sprintf("%s!C:D", sheetName) 
		
		resp, err := c.Service.Spreadsheets.Values.Get(c.SuratSpreadsheetID, readRange).Do()
		if err != nil {
			log.Printf("[SHEETS-WARN] Gagal membaca tab %s: %v", sheetName, err)
			continue // Coba sheet berikutnya
		}

		for _, row := range resp.Values {
			if len(row) >= 2 {
				idCellValue, idOk := row[0].(string)     // Kolom C (index 0)
				statusCellValue, statusOk := row[1].(string) // Kolom D (index 1)

				if idOk && statusOk && strings.EqualFold(idCellValue, unikID) { // Gunakan EqualFold (case-insensitive)
					return statusCellValue, nil // Ditemukan!
				}
			}
		}
	}
	
	return "", fmt.Errorf("ID Surat tidak ditemukan")
}
