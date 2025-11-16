package gemini

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type GeminiService struct {
	client         *GeminiClient
	villageContext string
}

// InfoDesa struct untuk mem-parsing info_desa.json
type InfoDesa struct {
	NamaDesa       string `json:"nama_desa"`
	Kecamatan      string `json:"kecamatan"`
	Kabupaten      string `json:"kabupaten"`
	JamOperasional string `json:"jam_operasional"`
	PerangkatDesa  []struct {
		Jabatan string `json:"jabatan"`
		Nama    string `json:"nama"`
	} `json:"perangkat_desa"`
}

func NewGeminiService(apiKey string) *GeminiService {
	client := NewGeminiClient(apiKey)
	context := loadVillageContext("knowledge/info_desa.json")
	return &GeminiService{
		client:         client,
		villageContext: context,
	}
}

func loadVillageContext(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("[WARN] Gagal memuat %s: %v. Gemini tidak akan memiliki konteks desa.", filePath, err)
		return "Informasi tentang desa tidak tersedia."
	}

	var info InfoDesa
	if err := json.Unmarshal(data, &info); err != nil {
		log.Printf("[WARN] Gagal mem-parsing %s: %v. Konteks tidak akan dimuat.", filePath, err)
		return "Informasi tentang desa tidak tersedia (format JSON salah)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Nama Desa: %s, Kecamatan: %s, Kabupaten: %s.\n", info.NamaDesa, info.Kecamatan, info.Kabupaten))
	sb.WriteString(fmt.Sprintf("Jam operasional kantor: %s.\n", info.JamOperasional))
	sb.WriteString("Perangkat Desa:\n")
	for _, p := range info.PerangkatDesa {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", p.Jabatan, p.Nama))
	}
	log.Println("[INFO] Konteks info_desa.json berhasil dimuat untuk Gemini.")
	return sb.String()
}

// HandleGeminiPrompt adalah satu-satunya fungsi yang dipanggil oleh router
func (s *GeminiService) HandleGeminiPrompt(question string) string {
	systemPrompt := fmt.Sprintf(`Kamu adalah 'SAS-BOT', asisten AI resmi Desa Sindang Anom.
Peranmu ada dua:
1.  **Menjawab Pertanyaan:** Jika pengguna bertanya tentang informasi desa (misal: "siapa kades?", "jam buka"), jawab berdasarkan KONTEKS di bawah.
2.  **Mengarahkan ke Menu:** Jika pengguna hanya menyapa (misal: "halo", "hi", "p") atau meminta bantuan umum (misal: "tolong", "bantu"), JANGAN menjawab pertanyaannya. Balas dengan sapaan ramah dan arahkan mereka untuk menggunakan menu utama.

Gaya Bahasa: Gunakan bahasa Indonesia yang ramah, sopan, dan dinamis. Hindari kata 'Tentu'.

--- KONTEKS INFORMASI DESA ---
%s
--- AKHIR KONTEKS ---
`, s.villageContext)

	response, err := s.client.GenerateContent(systemPrompt, question)
	if err != nil {
		log.Printf("[GEMINI-ERROR] %v", err)
		return "Maaf, sedang ada gangguan teknis pada layanan AI kami."
	}
	return response
}
