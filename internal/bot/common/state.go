// file: internal/bot/common/state.go
package common

// --- State Awal (Pre-Login) ---
const (
	STEP_AWAL_WAIT_NIK      = 1 // (Default) Menunggu user memasukkan NIK
	STEP_AWAL_NIK_NOT_FOUND = 2 // NIK tidak ada, menunggu pilihan (1. Ulangi, 2. Daftar)
	STEP_AWAL_MENU_UTAMA    = 3 // NIK valid, "login", menunggu pilihan menu (1,2,3)
)

// --- State Alur Data Diri (Mulai dari 200) ---
const (
	STEP_MENU_DATA_DIRI       = 200 // Menu (1. Input, 2. Edit)
	STEP_EDIT_CARI_NIK        = 201 // Menunggu NIK untuk diedit
	STEP_NIK_DUPLIKATE        = 202 // Menunggu (edit nik / stop)
	STEP_CHECKPOINT_SUKU      = 203 // Menunggu (cukup / lanjut)
	STEP_KONFIRMASI_DATA_DIRI = 204 // Menunggu (valid / edit)
	STEP_EDIT_DATA_DIRI       = 205 // Dalam mode edit (menunggu nomor 1-39)
	STEP_ULASAN_DATA_DIRI     = 206 // Menunggu ulasan (1-5)
)

// --- State Alur Surat (Mulai dari 500) ---
const (
	STEP_SURAT_MENU_UTAMA       = 500
	STEP_SURAT_VALIDASI_NIK     = 501
	STEP_SURAT_PILIH_JENIS      = 502
	STEP_SURAT_INPUT_DATA       = 503
	STEP_SURAT_KONFIRMASI_FIELD = 504
	STEP_SURAT_CEK_PROGRES      = 505
	STEP_ULASAN_SURAT           = 506
)

// --- State Alur Pengaduan (Mulai dari 800) ---
const (
	STEP_PENGADUAN_MENU         = 800
	STEP_PENGADUAN_VALIDASI_NIK = 801
	STEP_PENGADUAN_CARI_ID      = 802
	STEP_PENGADUAN_WAITING      = 803
	STEP_ULASAN_PENGADUAN       = 804
)
