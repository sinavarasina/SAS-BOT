package gemini

import (
	"fmt"
	"log"
	"strings"
)

type GeminiService struct {
	client *GeminiClient
}

//punya rey sebelumnya digunakan disini, jadi ga pake json lagi 
const villageContext = `
KONTEKS DESA SINDANG ANOM:
- Lokasi: Kec. Sekampung Udik, Kab. Lampung Timur.
- Kepala Desa: Aminudin.
- Jam Kantor: Senin - Jumat, 08:00 - 15:00 WIB.
- Kontak: 081368774938 / desaindang.anom@gmail.com.

LAYANAN SURAT:
1. Pengantar KTP/KK (Gratis, 1-2 hari)
2. Ket. Usaha (Bayar, 3-5 hari)
3. SKTM (Gratis, 2-7 hari)
4. Domisili (Bayar, 1-2 hari)
5. Kematian (Gratis, Hari sama)
6. Kelahiran (Gratis, 1-3 hari)

MENU UTAMA BOT:
1. Data Diri
2. Pengajuan Surat
3. Pengaduan
`

func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{
		client: NewGeminiClient(apiKey),
	}
}

func (s *GeminiService) HandleGeminiPrompt(userText string) string {
	maxInputChars := 2000
	if len(userText) > maxInputChars {
		userText = userText[:maxInputChars] + "...(dipotong)"
	}

	systemPrompt := fmt.Sprintf(`%s

PERAN & INSTRUKSI:
Kamu adalah 'SAS-BOT', asisten AI resmi Desa Sindang Anom. Tugasmu hanya dua:

1. JIKA USER MENYAPA / MINTA BANTUAN UMUM (misal: "halo", "p", "tolong", "menu"):
   - JANGAN coba menjawab atau berhalusinasi.
   - Cukup sapa balik dengan ramah dan arahkan mereka untuk mengetik angka menu (1, 2, atau 3).

2. JIKA USER BERTANYA INFO DESA (misal: "siapa kades?", "jam buka", "syarat sktm"):
   - Jawab LANGSUNG dan SINGKAT berdasarkan data KONTEKS di atas.
   - Jika jawaban tidak ada di konteks, minta maaf dan sarankan hubungi kontak kantor.

GAYA BAHASA:
- Gunakan Bahasa Indonesia yang luwes, sopan, dan menggunakan emoji.
- HARAM menggunakan kata "Tentu" atau "Baiklah" di awal kalimat.
- Maksimal jawaban 2-3 kalimat.`, villageContext)

	// Eksekusi request ke Gemini
	response, err := s.client.GenerateContent(systemPrompt, userText)
	if err != nil {
		log.Printf("[GEMINI-ERROR] %v", err)
		return "Maaf, server AI sedang sibuk. 😓 Silakan ketik angka menu (1, 2, atau 3) untuk melanjutkan manual."
	}

	return strings.TrimSpace(response)
}
