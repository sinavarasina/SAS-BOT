package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GeminiClient struct {
	APIKey string
	URL    string
}

func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		APIKey: apiKey,
		URL:    fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey),
	}
}

func (c *GeminiClient) GenerateContent(systemPrompt, userQuestion string) (string, error) {
	if c.APIKey == "" {
		return "Maaf, fitur AI sedang nonaktif (API Key tidak ada).", nil
	}

	fullPrompt := systemPrompt + "\n\n---\n\nPERTANYAAN USER:\n" + userQuestion

	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gagal marshal request Gemini: %w", err)
	}

	//timeout bot, while gemini server lagging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(context.TODO(), "POST", c.URL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("gagal membuat request Gemini: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("gagal request ke Gemini: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respons Gemini: %w", err)
	}

	// Struct untuk unmarshal respons
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Code		int		 `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[GEMINI-ERROR] Gagal unmarshal: %v. Respons: %s", err, string(respBody))
		return "", fmt.Errorf("gagal unmarshal respons Gemini: %w", err)
	}

	if result.Error.Message != "" {
    log.Printf("[GEMINI-API-ERROR] Code: %d, Status: %s, Message: %s", result.Error.Code, result.Error.Status, result.Error.Message)
    
    // Cek kode error 429 (Too Many Requests) atau resource exhausted
    if result.Error.Code == 429 || result.Error.Status == "RESOURCE_EXHAUSTED" {
        return "⏳ Server AI sedang sibuk (antrian penuh). Mohon tunggu 1 menit lalu tanya lagi.", nil
    }
    
    return "Maaf, sistem AI sedang gangguan teknis.", nil
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	log.Printf("[GEMINI-WARN] Respons AI kosong atau tidak valid: %s", string(respBody))
	return "Maaf, terjadi kesalahan saat memproses jawaban. 🤖", nil
}
