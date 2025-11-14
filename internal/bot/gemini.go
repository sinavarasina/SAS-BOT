package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultGeminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Hardcoded context data dari homechat.json untuk response yang lebih cerdas
var villageContext = `
KONTEKS DESA SINDANG ANOM:
- Nama Desa: Sindang Anom
- Kecamatan: Sekampung Udik
- Kabupaten: Lampung Timur
- Kepala Desa: Aminudin
- Jumlah Penduduk: ± 5.000 jiwa
- Luas Wilayah: 1.343 Ha
- Potensi Utama: Pertanian (Padi, Jagung) dan Perkebunan (Karet, Sawit)
- Kontak Kantor: 081368774938 / desaindang.anom@gmail.com
- Jam Operasional: Senin-Sabtu, 08:00-15:00 WIB

LAYANAN SURAT YANG TERSEDIA:
1. Surat Pengantar KTP/KK (Gratis, 1-2 hari)
2. Surat Keterangan Usaha (Rp. 50-100rb, 3-5 hari)
3. Surat Keterangan Tidak Mampu (Gratis, 2-7 hari)
4. Surat Keterangan Domisili (Gratis-50rb, 1-2 hari)
5. Surat Kematian (Gratis-25rb, Hari yang sama)
6. Surat Kelahiran (Gratis, 1-3 hari)

FITUR CHATBOT:
- Menu 1: Data Diri (Input/Edit data pribadi)
- Menu 2: Pengajuan Surat (Buat surat atau cek progres)
- Menu 3: Pengaduan (Lapor masalah atau saran)

PANDUAN RESPONS:
- Gunakan emoji yang relevan
- Maksimal 2-3 kalimat
- Hindari mengulang menu penuh
- Jangan gunakan kata "Tentu"
- Arahkan ke menu yang sesuai dengan pertanyaan user
- Jika user bertanya tentang surat/layanan, sebutkan biaya dan waktu proses
- Jika user bertanya tentang contact/jam, sebutkan informasi dari konteks
- Pastikan respons friendly dan helpful
`

func HandleGeminiPrompt(userText string) string {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	apiURL := strings.TrimSpace(os.Getenv("GEMINI_API_URL"))
	if apiURL == "" {
		apiURL = defaultGeminiURL
	}

	if apiKey == "" {
		go log.Printf("[AI-WARN] GEMINI_API_KEY is empty; returning generic response")
		return "_👋 Halo! Ada yang bisa saya bantu? Silakan pilih menu untuk melanjutkan._"
	}

	// Channel untuk menerima response dari goroutine
	respChan := make(chan string, 1)

	// Jalankan semua proses di goroutine
	go func() {
		prompt := fmt.Sprintf(
			"%s\n\n"+
				"Pengguna mengirim: '%s'\n\n"+
				"Berdasarkan konteks di atas, balas dengan 1-2 kalimat yang ramah, gunakan emoji, "+
				"dan arahkan ke menu yang sesuai (Menu 1 untuk data diri, Menu 2 untuk surat, Menu 3 untuk pengaduan). "+
				"Jika user bertanya tentang jam kerja/kontak, sebutkan jam operasional dan nomor yang ada. "+
				"Hindari mengulang teks menu penuh dan jangan gunakan kata 'Tentu'. "+
				"Berikan respons yang spesifik dan helpful sesuai pertanyaan mereka. "+
				"Gunakan emoji untuk membuat respons lebih menarik dan interaktif.",
			villageContext,
			userText,
		)

		req := geminiRequest{
			Contents: []struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			}{
				{Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				}},
			},
		}

		payload, err := json.Marshal(req)
		if err != nil {
			go log.Printf("[AI-ERROR] marshal request: %v", err)
			respChan <- "_👋 Maaf, sedang sibuk. Silakan coba lagi._"
			return
		}

		url := fmt.Sprintf("%s?key=%s", apiURL, apiKey)
		httpClient := &http.Client{Timeout: 15 * time.Second}
		resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			go log.Printf("[AI-ERROR] POST request failed: %v", err)
			respChan <- "_👋 Maaf, sedang sibuk. Silakan coba lagi._"
			return
		}
		defer resp.Body.Close()

		go log.Printf("[AI-DEBUG] Response Status: %d", resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			go log.Printf("[AI-ERROR] read body: %v", err)
			respChan <- "_👋 Maaf, sedang sibuk. Silakan coba lagi._"
			return
		}

		var gr geminiResponse
		if err := json.Unmarshal(body, &gr); err != nil {
			go func() {
				log.Printf("[AI-ERROR] unmarshal response: %v", err)
				log.Printf("[AI-DEBUG] Response Body: %s", string(body))
			}()
			respChan <- "_👋 Maaf, sedang sibuk. Silakan coba lagi._"
			return
		}

		// CHECK CANDIDATES PALING AWAL - return langsung jika valid
		if len(gr.Candidates) > 0 && len(gr.Candidates[0].Content.Parts) > 0 {
			responseText := strings.TrimSpace(gr.Candidates[0].Content.Parts[0].Text)

			if !strings.HasPrefix(responseText, "_") {
				responseText = "_" + responseText + "_"
			}

			// LANGSUNG kirim ke channel SEBELUM logging
			respChan <- responseText
			
			// Logging di background - non-blocking
			go func(rt string, respBody []byte) {
				log.Printf("[AI-SUCCESS] Response sent to client: %s", rt)
				log.Printf("[AI-DEBUG] Response Body: %s", string(respBody))
			}(responseText, body)
			
			return
		}

		go func() {
			log.Printf("[AI-ERROR] No candidates or parts in response. Candidates: %d", len(gr.Candidates))
			log.Printf("[AI-DEBUG] Response Body: %s", string(body))
		}()
		respChan <- "_👋 Silakan gunakan menu untuk melanjutkan._"
	}()

	// TUNGGU response dari channel dengan timeout 10 detik
	select {
	case resp := <-respChan:
		return resp
	case <-time.After(10 * time.Second):
		return "_⏳ Sedang memproses..._"
	}
}