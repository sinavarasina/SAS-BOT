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

	fullPrompt := systemPrompt + "\n\n---\n\n" + userQuestion

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
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
		// Menambahkan penanganan error dari API
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[GEMINI-ERROR] Gagal unmarshal: %v. Respons: %s", err, string(respBody))
		return "", fmt.Errorf("gagal unmarshal respons Gemini: %w", err)
	}

	if result.Error.Message != "" {
		log.Printf("[GEMINI-WARN] AI tidak memberikan jawaban. Respons: %s", string(respBody))
		return "Maaf, saya tidak bisa memproses permintaan Anda saat ini. 😥", nil
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	log.Printf("[GEMINI-WARN] Respons AI kosong atau tidak valid: %s", string(respBody))
	return "Maaf, terjadi kesalahan saat memproses jawaban. 🤖", nil
}
