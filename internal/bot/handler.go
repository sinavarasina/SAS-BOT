package bot

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string) string {
	text = strings.TrimSpace(text)

	WAUser := db.User{
		JID:       jid,
		Number:    number,
		Username:  username,
		Previlege: "user",
	}

	err := db.SaveUser(dbConn, WAUser)
	if err != nil {
		log.Printf("Error at db.SaveUser for jid: %s, Message : %v", jid, err)
	}
	
	//sesi pengaduan
	s := GetSession(jid)
	if s.Step == "menunggu_pengaduan" {
		return "Mohon kirimkan gambar beserta deskripsi pengaduan Anda."
	}
	
	switch text {
	case "!batal":
		ResetSession(jid)
		return "Sesi Dibatalkan"
		case "2":
		s.Step = "menunggu_pengaduan"
		return "Anda memilih buat pengaduan.\n\nSilakan kirimkan *satu foto* pengaduan Anda, dan *tulis deskripsi* di bagian caption/keterangan gambar tersebut."
	}

	// Semua pesan dilempar ke Gemini, materi dikosongkan
	materi := "Materi: "
	pertanyaan := "Pertanyaan: " + text
	geminiResp, err := askGemini(materi, pertanyaan)
	if err != nil {
		log.Printf("Error Gemini API: %v", err)
		return "Maaf, terjadi kesalahan saat memproses permintaan Anda."
	}
	return geminiResp
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	// Hapus pesan template, bisa return kosong atau pesan singkat lain jika diinginkan
	return ""
}

// askGemini mengirim permintaan ke Gemini API dan mengembalikan responnya sebagai string.
func askGemini(materi, pertanyaan string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=AIzaSyCR3Weo_JBPE_PnWNLEfo4T57Uw0bqCQM4"

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": materi},
					{"text": pertanyaan},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parsing response Gemini
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
		return "", err
	}
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "Maaf, tidak ada respon dari AI.", nil
}
