package datadiri

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidateInput memeriksa input pengguna untuk setiap langkah
func ValidateInput(text string, step Step) (interface{}, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("Input tidak boleh kosong\n\n%s", FormatQuestion(step))
	}

	if step.IsDate {
		t, err := time.Parse("02-01-2006", text)
		if err != nil {
			return nil, fmt.Errorf("⚠️ Format tanggal salah. Harap masukkan dengan format DD-MM-YYYY (contoh: 25-12-2024)\n\n%s", FormatQuestion(step))
		}
		return t, nil
	}

	if step.Options != nil {
		choice, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("⚠️ Mohon pilih nomor yang tersedia dari daftar di bawah\n\n%s", FormatQuestion(step))
		}
		if _, valid := step.Options[choice]; !valid {
			return nil, fmt.Errorf("⚠️ Pilihan tidak valid. Mohon pilih nomor yang tersedia dari daftar di bawah\n\n%s", FormatQuestion(step))
		}
		return choice, nil
	}

	switch step.Field {
	case "nik", "no_kk":
		if len(text) != 16 {
			return nil, fmt.Errorf("⚠️ %s harus 16 digit\nPanjang input saat ini: %d digit\n\n%s",
				strings.ToLower(step.Field), len(text), FormatQuestion(step))
		}
		if _, err := strconv.ParseInt(text, 10, 64); err != nil {
			return nil, fmt.Errorf("⚠️ %s hanya boleh berisi angka\n\n%s",
				strings.ToLower(step.Field), FormatQuestion(step))
		}
	case "rt":
		if len(text) != 3 {
			return nil, fmt.Errorf("⚠️ RT harus 3 digit\nContoh format yang benar: 001\n\n%s", step.Question)
		}
		if _, err := strconv.Atoi(text); err != nil {
			return nil, fmt.Errorf("⚠️ RT hanya boleh berisi angka\n\n%s", step.Question)
		}
	case "nama", "nama_ayah", "nama_ibu":
		if len(text) < 2 {
			return nil, fmt.Errorf("⚠️ %s terlalu pendek\nMinimal 2 karakter\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
		if strings.ContainsAny(text, "0123456789!@#$%^&*()_+=[]{};:'\"\\|,.<>/?") {
			return nil, fmt.Errorf("⚠️ %s tidak boleh mengandung angka atau karakter khusus\n\n%s",
				strings.ToLower(step.Field), step.Question)
		}
	case "tempat_lahir":
		if len(text) < 3 {
			return nil, fmt.Errorf("⚠️ Tempat lahir terlalu pendek\nMinimal 3 karakter\n\n%s",
				step.Question)
		}
	}

	return text, nil
}
