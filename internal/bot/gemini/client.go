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
		URL:    fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey),
	}
}

func (c *GeminiClient) GenerateContent(systemPrompt, userQuestion string) (string, error) {
	if c.APIKey == "" {
		return "Maaf, fitur AI sedang nonaktif (API Key tidak ada).", nil
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": systemPrompt}}},
			{"parts": []map[string]string{{"text": userQuestion}}},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gagal marshal request Gemini: %w", err)
	}

	resp, err := http.Post(c.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("gagal request ke Gemini: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respons Gemini: %w", err)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[GEMINI-ERROR] Gagal unmarshal: %v. Respons: %s", err, string(respBody))
		return "Maaf, format respons dari AI tidak valid.", nil
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	
	log.Printf("[GEMINI-WARN] AI tidak memberikan jawaban. Respons: %s", string(respBody))
	return "Maaf, AI tidak memberikan jawaban. Coba tanyakan hal lain.", nil
}
