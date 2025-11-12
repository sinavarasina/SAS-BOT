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

func HandleGeminiPrompt(userText string) string {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	apiURL := strings.TrimSpace(os.Getenv("GEMINI_API_URL"))
	if apiURL == "" {
		apiURL = defaultGeminiURL
	}

	if apiKey == "" {
		log.Printf("[AI-WARN] GEMINI_API_KEY is empty; returning generic response")
		return "👋 Halo! Ada yang bisa saya bantu? Silakan pilih menu untuk melanjutkan."
	}

	prompt := fmt.Sprintf(
		"Kamu adalah bot asisten untuk pencatatan data penduduk di desa Sindang Anom. "+
			"Pengguna mengirim: '%s'. "+
			"Balas dalam 1-2 kalimat yang ramah dengan menggunakan emoji yang relevan, dan arahkan agar memilih menu yang sesuai (1 untuk data diri, 2 untuk pengajuan surat, 3 untuk pengaduan). "+
			"Hindari mengulang teks menu penuh dan jangan gunakan kata 'Tentu'. "+
			"Gunakan emoji untuk membuat respons lebih menarik dan interaktif.",
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
		log.Printf("[AI-ERROR] marshal request: %v", err)
		return "👋 Maaf, sedang sibuk. Silakan coba lagi."
	}

	url := fmt.Sprintf("%s?key=%s", apiURL, apiKey)
	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("[AI-ERROR] POST request: %v", err)
		return "👋 Maaf, sedang sibuk. Silakan coba lagi."
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AI-ERROR] read body: %v", err)
		return "👋 Maaf, sedang sibuk. Silakan coba lagi."
	}

	var gr geminiResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		log.Printf("[AI-ERROR] unmarshal response: %v", err)
		return "👋 Maaf, sedang sibuk. Silakan coba lagi."
	}

	if len(gr.Candidates) > 0 && len(gr.Candidates[0].Content.Parts) > 0 {
		return strings.TrimSpace(gr.Candidates[0].Content.Parts[0].Text)
	}
	return "👋 Silakan gunakan menu untuk melanjutkan."
}

