package datadiri

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/bot/common"
	"github.com/sinavarasina/SAS-BOT/internal/db"
	"github.com/sinavarasina/SAS-BOT/internal/sheets"
)

// Service menangani logika bisnis untuk modul Data Diri
type Service struct {
	DB           *sqlx.DB
	SheetsClient *sheets.SheetsClient
}

func NewService(ctx *common.ServiceContext) *Service {
	return &Service{
		DB:           ctx.DB,
		SheetsClient: ctx.SheetsClient,
	}
}

// CheckNIKExists memeriksa apakah NIK sudah ada
func (s *Service) CheckNIKExists(nik, jid string) (string, error) {
	isDuplicate, err := db.CheckNIKExistsInPenduduk(s.DB, nik)
	if err != nil {
		log.Printf("[ERROR] Gagal mengecek NIK di DB permanen: %v", err)
		return "Maaf, terjadi kesalahan sistem saat validasi NIK (DB).", err
	}
	if isDuplicate {
		return "⚠️ NIK ini sudah terdaftar di data penduduk desa.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran.", fmt.Errorf("nik duplikat permanen")
	}

	isDuplicate, err = db.CheckNIKExists(s.DB, nik, jid)
	if err != nil {
		log.Printf("[ERROR] Gagal mengecek NIK di DB sesi: %v", err)
		return "Maaf, terjadi kesalahan sistem saat validasi NIK (Sesi).", err
	}
	if isDuplicate {
		return "⚠️ NIK ini sedang didaftarkan oleh pengguna lain saat ini.\n\nKetik 'edit nik' untuk memasukkan NIK baru, atau 'stop' untuk membatalkan pendaftaran.", fmt.Errorf("nik duplikat sesi")
	}

	return "", nil
}

// SaveDataToDBAndSheets menyimpan data akhir
func (s *Service) SaveDataToDBAndSheets(jid string) error {
	fullSession, err := db.GetFullSessionData(s.DB, jid)
	if err != nil {
		log.Printf("[ERROR] Gagal mengambil data lengkap: %v", err)
		return err
	}

	if err := db.SaveDataPenduduk(s.DB, *fullSession); err != nil {
		return fmt.Errorf("maaf, terjadi kesalahan besar saat menyimpan ke database: %v", err)
	}

	go func() {
		nik := fullSession.NIK.String
		rowNum, err := s.SheetsClient.FindRowByNIK(nik)
		if err != nil {
			log.Printf("Menambah NIK %s baru ke Sheet.", nik)
			s.SheetsClient.AppendDataPenduduk(*fullSession)
		} else {
			log.Printf("Mengupdate NIK %s di baris %d Sheet.", nik, rowNum)
			s.SheetsClient.UpdateRowData(rowNum, *fullSession)
		}
	}()

	return nil
}
