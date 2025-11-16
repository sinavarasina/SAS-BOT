package surat

import (
	"fmt"
	"log"
	"os"
)

func (s *Service) NotifyPerangkatDesa(pdfPath, namaSurat, uniqueID, namaPengaju string) {
	nomorKades := os.Getenv("NOMOR_PERANGKAT_DESA")
	if nomorKades == "" {
		return
	}

	kadesJID := fmt.Sprintf("%s@s.whatsapp.net", nomorKades)
	log.Printf("[SURAT] Mengirim salinan PDF ke Perangkat Desa (%s)", kadesJID)
	
	caption := fmt.Sprintf(
		"Laporan Surat Baru Dibuat:\n\nJenis: *%s*\nID Unik: *%s*\nAtas Nama: *%s*",
		namaSurat, uniqueID, namaPengaju,
	)
	
	if err := SendFile(s.Ctx.WAClient, kadesJID, pdfPath, caption); err != nil {
		log.Printf("[SURAT-ERROR] Gagal mengirim file PDF ke Kades: %v", err)
	}
}
